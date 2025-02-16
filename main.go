package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"github.com/warthog618/go-gpiocdev/device/rpi"
)

var (
	DEV_HID = "/dev/hidg0"
	CHIP    = "/dev/gpiochip0"

	// Key code reference: https://gist.github.com/MightyPork/6da26e382a7ad91b5496ee55fdc73db2
	// todo 'g' key issue is addressed in the comments

	/**
	 * Modifier masks - used for the first byte in the HID report.
	 * NOTE: The second byte in the report is reserved, 0x00
	 */
	MOD_LCTRL  = byte(0x01)
	MOD_LSHIFT = byte(0x02)
	MOD_LALT   = byte(0x04)
	MOD_LMETA  = byte(0x08)
	MOD_RCTRL  = byte(0x10)
	MOD_RSHIFT = byte(0x20)
	MOD_RALT   = byte(0x40)
	MOD_RMETA  = byte(0x80)

	/**
	 * Scan codes - last N slots in the HID report (usually 6).
	 * 0x00 if no key pressed.
	 *
	 * If more than N keys are pressed, the HID reports
	 * KEY_ERR_OVF in all slots to indicate this condition.
	 */

	KEY_NONE    = byte(0x00) // No key pressed
	KEY_ERR_OVF = byte(0x01) //  Keyboard Error Roll Over - used for all slots if too many keys are pressed ("Phantom key")
	// 0x02 //  Keyboard POST Fail
	// 0x03 //  Keyboard Error Undefined

	A           = byte(0x04) // Keyboard a and A
	KEY_UPPER_A = []byte{MOD_LSHIFT, 0x00, A, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_A = []byte{KEY_NONE, 0x00, A, 0x00, 0x00, 0x00, 0x00, 0x00}

	B           = byte(0x05) // Keyboard b and B
	KEY_UPPER_B = []byte{MOD_LSHIFT, 0x00, B, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_B = []byte{KEY_NONE, 0x00, B, 0x00, 0x00, 0x00, 0x00, 0x00}

	C           = byte(0x06) // Keyboard c and C
	KEY_UPPER_C = []byte{MOD_LSHIFT, 0x00, C, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_C = []byte{KEY_NONE, 0x00, C, 0x00, 0x00, 0x00, 0x00, 0x00}

	D           = byte(0x07) // Keyboard d and D
	KEY_UPPER_D = []byte{MOD_LSHIFT, 0x00, D, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_D = []byte{KEY_NONE, 0x00, D, 0x0D, 0x00, 0x00, 0x00, 0x00}

	E           = byte(0x08) // Keyboard e and E
	KEY_UPPER_E = []byte{MOD_LSHIFT, 0x00, E, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_E = []byte{KEY_NONE, 0x00, E, 0x0D, 0x00, 0x00, 0x00, 0x00}

	F           = byte(0x09) // Keyboard f and F
	KEY_UPPER_F = []byte{MOD_LSHIFT, 0x00, F, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_F = []byte{KEY_NONE, 0x00, F, 0x0D, 0x00, 0x00, 0x00, 0x00}

	G           = byte(0x0a) // Keyboard g and G
	KEY_UPPER_G = []byte{MOD_LSHIFT, 0x00, G, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_G = []byte{KEY_NONE, 0x00, G, 0x0D, 0x00, 0x00, 0x00, 0x00}

	H           = byte(0x0b) // Keyboard h and H
	KEY_UPPER_H = []byte{MOD_LSHIFT, 0x00, H, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_H = []byte{KEY_NONE, 0x00, H, 0x0D, 0x00, 0x00, 0x00, 0x00}

	I           = byte(0x0c) // Keyboard i and I
	KEY_UPPER_I = []byte{MOD_LSHIFT, 0x00, I, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_I = []byte{KEY_NONE, 0x00, I, 0x0D, 0x00, 0x00, 0x00, 0x00}

	J           = byte(0x0d) // Keyboard j and J
	KEY_UPPER_J = []byte{MOD_LSHIFT, 0x00, J, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_J = []byte{KEY_NONE, 0x00, J, 0x0D, 0x00, 0x00, 0x00, 0x00}

	K           = byte(0x0e) // Keyboard k and K
	KEY_UPPER_K = []byte{MOD_LSHIFT, 0x00, K, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_K = []byte{KEY_NONE, 0x00, K, 0x0D, 0x00, 0x00, 0x00, 0x00}

	L           = byte(0x0f) // Keyboard l and L
	KEY_UPPER_L = []byte{MOD_LSHIFT, 0x00, L, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_L = []byte{KEY_NONE, 0x00, L, 0x0D, 0x00, 0x00, 0x00, 0x00}

	M           = byte(0x10) // Keyboard m and M
	KEY_UPPER_M = []byte{MOD_LSHIFT, 0x00, M, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_M = []byte{KEY_NONE, 0x00, M, 0x0D, 0x00, 0x00, 0x00, 0x00}

	N           = byte(0x11) // Keyboard n and N
	KEY_UPPER_N = []byte{MOD_LSHIFT, 0x00, N, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_N = []byte{KEY_NONE, 0x00, N, 0x0D, 0x00, 0x00, 0x00, 0x00}

	O           = byte(0x12) // Keyboard o and O
	KEY_UPPER_O = []byte{MOD_LSHIFT, 0x00, O, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_O = []byte{KEY_NONE, 0x00, O, 0x0D, 0x00, 0x00, 0x00, 0x00}

	P           = byte(0x13) // Keyboard p and P
	KEY_UPPER_P = []byte{MOD_LSHIFT, 0x00, P, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_P = []byte{KEY_NONE, 0x00, P, 0x0D, 0x00, 0x00, 0x00, 0x00}

	Q           = byte(0x14) // Keyboard q and Q
	KEY_UPPER_Q = []byte{MOD_LSHIFT, 0x00, Q, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_Q = []byte{KEY_NONE, 0x00, Q, 0x0D, 0x00, 0x00, 0x00, 0x00}

	R           = byte(0x15) // Keyboard r and R
	KEY_UPPER_R = []byte{MOD_LSHIFT, 0x00, R, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_R = []byte{KEY_NONE, 0x00, R, 0x0D, 0x00, 0x00, 0x00, 0x00}

	S           = byte(0x16) // Keyboard s and S
	KEY_UPPER_S = []byte{MOD_LSHIFT, 0x00, S, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_S = []byte{KEY_NONE, 0x00, S, 0x0D, 0x00, 0x00, 0x00, 0x00}

	T           = byte(0x17) // Keyboard t and T
	KEY_UPPER_T = []byte{MOD_LSHIFT, 0x00, T, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_T = []byte{KEY_NONE, 0x00, T, 0x0D, 0x00, 0x00, 0x00, 0x00}

	U           = byte(0x18) // Keyboard u and U
	KEY_UPPER_U = []byte{MOD_LSHIFT, 0x00, U, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_U = []byte{KEY_NONE, 0x00, U, 0x0D, 0x00, 0x00, 0x00, 0x00}

	V           = byte(0x19) // Keyboard v and V
	KEY_UPPER_V = []byte{MOD_LSHIFT, 0x00, V, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_V = []byte{KEY_NONE, 0x00, V, 0x0D, 0x00, 0x00, 0x00, 0x00}

	W           = byte(0x1a) // Keyboard w and W
	KEY_UPPER_W = []byte{MOD_LSHIFT, 0x00, W, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_W = []byte{KEY_NONE, 0x00, W, 0x0D, 0x00, 0x00, 0x00, 0x00}

	X           = byte(0x1b) // Keyboard x and X
	KEY_UPPER_X = []byte{MOD_LSHIFT, 0x00, X, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_X = []byte{KEY_NONE, 0x00, X, 0x0D, 0x00, 0x00, 0x00, 0x00}

	Y           = byte(0x1c) // Keyboard y and Y
	KEY_UPPER_Y = []byte{MOD_LSHIFT, 0x00, Y, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_Y = []byte{KEY_NONE, 0x00, Y, 0x0D, 0x00, 0x00, 0x00, 0x00}

	Z           = byte(0x1d) // Keyboard z and Z
	KEY_UPPER_Z = []byte{MOD_LSHIFT, 0x00, Z, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_LOWER_Z = []byte{KEY_NONE, 0x00, Z, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_1                    = byte(0x1e) // Keyboard 1 and !
	KEY_EXCLAMATION_PIONT = []byte{MOD_LSHIFT, 0x00, _1, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_ONE               = []byte{KEY_NONE, 0x00, _1, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_2            = byte(0x1f) // Keyboard 2 and @
	KEY_AT_SYMBOL = []byte{MOD_LSHIFT, 0x00, _2, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_TWO       = []byte{KEY_NONE, 0x00, _2, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_3              = byte(0x20) // Keyboard 3 and #
	KEY_HASH_SYMBOL = []byte{MOD_LSHIFT, 0x00, _3, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_THREE       = []byte{KEY_NONE, 0x00, _3, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_4              = byte(0x21) // Keyboard 4 and $
	KEY_DOLLAR_SIGN = []byte{MOD_LSHIFT, 0x00, _4, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_FOUR        = []byte{KEY_NONE, 0x00, _4, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_5                 = byte(0x22) // Keyboard 5 and %
	KEY_PERCENT_SYMBOL = []byte{MOD_LSHIFT, 0x00, _5, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_FIVE           = []byte{KEY_NONE, 0x00, _5, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_6        = byte(0x23) // Keyboard 6 and ^
	KEY_CARET = []byte{MOD_LSHIFT, 0x00, _6, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_SIX   = []byte{KEY_NONE, 0x00, _6, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_7             = byte(0x24) // Keyboard 7 and &
	KEY_AND_SYMBOL = []byte{MOD_LSHIFT, 0x00, _7, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_SEVEN      = []byte{KEY_NONE, 0x00, _7, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_8        = byte(0x25) // Keyboard 8 and *
	KEY_STAR  = []byte{MOD_LSHIFT, 0x00, _8, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_EIGHT = []byte{KEY_NONE, 0x00, _8, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_9             = byte(0x26) // Keyboard 9 and (
	KEY_OPEN_PAREN = []byte{MOD_LSHIFT, 0x00, _9, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_NINE       = []byte{KEY_NONE, 0x00, _9, 0x0D, 0x00, 0x00, 0x00, 0x00}

	_0              = byte(0x27) // Keyboard 0 and )
	KEY_CLOSE_PAREN = []byte{MOD_LSHIFT, 0x00, _0, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_ZERO        = []byte{KEY_NONE, 0x00, _0, 0x0D, 0x00, 0x00, 0x00, 0x00}

	KEY_ENTER     = byte(0x28) // Keyboard Return (ENTER)
	KEY_ESC       = byte(0x29) // Keyboard ESCAPE
	KEY_BACKSPACE = byte(0x2a) // Keyboard DELETE (Backspace)
	KEY_TAB       = byte(0x2b) // Keyboard Tab
	KEY_SPACE     = byte(0x2c) // Keyboard Spacebar

	MINUS           = byte(0x2d) // Keyboard - and _
	KEY_UNDER_SCORE = []byte{MOD_LSHIFT, 0x00, MINUS, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_MINUS       = []byte{KEY_NONE, 0x00, MINUS, 0x0D, 0x00, 0x00, 0x00, 0x00}

	EQUAL     = byte(0x2e) // Keyboard = and +
	KEY_PLUS  = []byte{MOD_LSHIFT, 0x00, EQUAL, 0x00, 0x00, 0x00, 0x00, 0x00}
	KEY_EQUAL = []byte{KEY_NONE, 0x00, EQUAL, 0x0D, 0x00, 0x00, 0x00, 0x00}

	LEFTBRACE            = byte(0x2f) // Keyboard [ and {
	KEY_LEFT_CURLY_BRACE = []byte{MOD_LSHIFT, 0x00, LEFTBRACE, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_LEFTBRACE        = []byte{KEY_NONE, 0x00, LEFTBRACE, 0x00, 0x00, 0x00, 0x00, 0x00}

	RIGHTBRACE            = byte(0x30) // Keyboard ] and }
	KEY_RIGHT_CURLY_BRACE = []byte{MOD_LSHIFT, 0x00, RIGHTBRACE, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_RIGHTBRACE        = []byte{KEY_NONE, 0x00, RIGHTBRACE, 0x00, 0x00, 0x00, 0x00, 0x00}

	BACKSLASH     = byte(0x31) // Keyboard \ and |
	KEY_PIPE      = []byte{MOD_LSHIFT, 0x00, BACKSLASH, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_BACKSLASH = []byte{KEY_NONE, 0x00, BACKSLASH, 0x00, 0x00, 0x00, 0x00, 0x00}

	HASHTILDE = byte(0x32) // Keyboard Non-US # and ~

	SEMICOLON     = byte(0x33) // Keyboard ; and :
	KEY_COLON     = []byte{MOD_LSHIFT, 0x00, SEMICOLON, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_SEMICOLON = []byte{KEY_NONE, 0x00, SEMICOLON, 0x00, 0x00, 0x00, 0x00, 0x00}

	APOSTROPHE        = byte(0x34) // Keyboard ' and "
	KEY_DOUBLE_QUOTES = []byte{MOD_LSHIFT, 0x00, APOSTROPHE, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_APOSTROPHE    = []byte{KEY_NONE, 0x00, APOSTROPHE, 0x00, 0x00, 0x00, 0x00, 0x00}

	GRAVE         = byte(0x35) // Keyboard ` and ~
	KEY_BACK_TICK = []byte{MOD_LSHIFT, 0x00, GRAVE, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_GRAVE     = []byte{KEY_NONE, 0x00, GRAVE, 0x00, 0x00, 0x00, 0x00, 0x00}

	COMMA                  = byte(0x36) // Keyboard , and <
	KEY_LEFT_ANGLE_BRACKET = []byte{MOD_LSHIFT, 0x00, COMMA, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_COMMA              = []byte{KEY_NONE, 0x00, COMMA, 0x00, 0x00, 0x00, 0x00, 0x00}

	DOT                     = byte(0x37) // Keyboard . and >
	KEY_RIGHT_ANGLE_BRACKET = []byte{MOD_LSHIFT, 0x00, DOT, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_PERIOD              = []byte{KEY_NONE, 0x00, DOT, 0x00, 0x00, 0x00, 0x00, 0x00}

	FORWARD_SLASH     = byte(0x38) // Keyboard / and ?
	KEY_QUESTION_MARK = []byte{MOD_LSHIFT, 0x00, FORWARD_SLASH, 0x0D, 0x00, 0x00, 0x00, 0x00}
	KEY_FORWARD_SLASH = []byte{KEY_NONE, 0x00, FORWARD_SLASH, 0x00, 0x00, 0x00, 0x00, 0x00}

	KEY_CAPSLOCK = byte(0x39) // Keyboard Caps Lock

	keyMap = map[rune][]byte{
		'a':  KEY_LOWER_A,
		'A':  KEY_UPPER_A,
		'b':  KEY_LOWER_B,
		'B':  KEY_UPPER_B,
		'c':  KEY_LOWER_C,
		'C':  KEY_UPPER_C,
		'd':  KEY_LOWER_D,
		'D':  KEY_UPPER_D,
		'e':  KEY_LOWER_E,
		'E':  KEY_UPPER_E,
		'f':  KEY_LOWER_F,
		'F':  KEY_UPPER_F,
		'g':  KEY_LOWER_G,
		'G':  KEY_UPPER_G,
		'h':  KEY_LOWER_H,
		'H':  KEY_UPPER_H,
		'i':  KEY_LOWER_I,
		'I':  KEY_UPPER_I,
		'j':  KEY_LOWER_J,
		'J':  KEY_UPPER_J,
		'k':  KEY_LOWER_K,
		'K':  KEY_UPPER_K,
		'l':  KEY_LOWER_L,
		'L':  KEY_UPPER_L,
		'm':  KEY_LOWER_M,
		'M':  KEY_UPPER_M,
		'n':  KEY_LOWER_N,
		'N':  KEY_UPPER_N,
		'o':  KEY_LOWER_O,
		'O':  KEY_UPPER_O,
		'p':  KEY_LOWER_P,
		'P':  KEY_UPPER_P,
		'q':  KEY_LOWER_Q,
		'Q':  KEY_UPPER_Q,
		'r':  KEY_LOWER_R,
		'R':  KEY_UPPER_R,
		's':  KEY_LOWER_S,
		'S':  KEY_UPPER_S,
		't':  KEY_LOWER_T,
		'T':  KEY_UPPER_T,
		'u':  KEY_LOWER_U,
		'U':  KEY_UPPER_U,
		'v':  KEY_LOWER_V,
		'V':  KEY_UPPER_V,
		'w':  KEY_LOWER_W,
		'W':  KEY_UPPER_W,
		'x':  KEY_LOWER_X,
		'X':  KEY_UPPER_X,
		'y':  KEY_LOWER_Y,
		'Y':  KEY_UPPER_Y,
		'z':  KEY_LOWER_Z,
		'Z':  KEY_UPPER_Z,
		'!':  KEY_EXCLAMATION_PIONT,
		'1':  KEY_ONE,
		'@':  KEY_AT_SYMBOL,
		'2':  KEY_TWO,
		'#':  KEY_HASH_SYMBOL,
		'3':  KEY_THREE,
		'$':  KEY_DOLLAR_SIGN,
		'4':  KEY_FOUR,
		'%':  KEY_PERCENT_SYMBOL,
		'5':  KEY_FIVE,
		'^':  KEY_CARET,
		'6':  KEY_SIX,
		'&':  KEY_AND_SYMBOL,
		'7':  KEY_SEVEN,
		'*':  KEY_STAR,
		'8':  KEY_EIGHT,
		'(':  KEY_OPEN_PAREN,
		'9':  KEY_NINE,
		')':  KEY_CLOSE_PAREN,
		'0':  KEY_ZERO,
		'_':  KEY_UNDER_SCORE,
		'-':  KEY_MINUS,
		'+':  KEY_PLUS,
		'=':  KEY_EQUAL,
		'{':  KEY_LEFT_CURLY_BRACE,
		'[':  KEY_LEFTBRACE,
		'}':  KEY_RIGHT_CURLY_BRACE,
		']':  KEY_RIGHTBRACE,
		'|':  KEY_PIPE,
		'\\': KEY_BACKSLASH,
		':':  KEY_COLON,
		';':  KEY_SEMICOLON,
		'"':  KEY_DOUBLE_QUOTES,
		'\'': KEY_APOSTROPHE,
		'`':  KEY_BACK_TICK,
		'~':  KEY_GRAVE,
		'<':  KEY_LEFT_ANGLE_BRACKET,
		',':  KEY_COMMA,
		'>':  KEY_RIGHT_ANGLE_BRACKET,
		'.':  KEY_PERIOD,
		'?':  KEY_QUESTION_MARK,
		'/':  KEY_FORWARD_SLASH,
	}
)

var d *debouncer
var f *os.File

func main() {
	log.Println("program started")
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()
	c, err := gpiocdev.NewChip(CHIP)
	if err != nil {
		log.Fatalf("error, when NewChip() for main(). Error: %v", err)
	}
	log.Println("chip added")

	d = newDebouncer(time.Second * 2)

	l, err := c.RequestLine(
		rpi.GPIO16,
		gpiocdev.WithEventHandler(handle16),
		gpiocdev.WithRisingEdge,
		gpiocdev.WithPullUp,
	)
	if err != nil {
		log.Fatalf("error, when RequestLine() for main(). Error: %v", err)
	}
	log.Println("event handler GPIO16 registered")
	defer l.Close()

	l2, err := c.RequestLine(
		rpi.GPIO25,
		gpiocdev.WithEventHandler(handle16),
		gpiocdev.WithRisingEdge,
		gpiocdev.WithPullUp,
	)
	if err != nil {
		log.Fatalf("error, when RequestLine() for main(). Error: %v", err)
	}
	log.Println("event handler GPIO25 egistered")
	defer l2.Close()

	f, err = os.OpenFile(DEV_HID, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening HID device:", err)
		return
	}
	log.Println("device file opened")

	defer f.Close()

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

type debouncer struct {
	delay   time.Duration
	timer   *time.Timer
	mu      sync.Mutex
	started bool
}

func newDebouncer(delay time.Duration) *debouncer {
	return &debouncer{
		delay: delay,
	}
}

func (d *debouncer) debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.started {
		return
	}
	d.started = true
	f()

	// Start a new timer that will call the function after the delay
	d.timer = time.AfterFunc(d.delay, func() {
		d.mu.Lock()       // we are locking before the function call because there could be multple
		d.started = false // Allow the next function call to use a new timer
		d.mu.Unlock()
	})
}

func handle16(evt gpiocdev.LineEvent) {
	d.debounce(func() {
		// Send multiple keys
		keys := []byte{A, B, C} // Typing 'abc'
		for _, key := range keys {
			if err := sendKey(f, key); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	})
}

func handle25(evt gpiocdev.LineEvent) {
	d.debounce(func() {
		fmt.Println("attempting to send keys!")
		// Send multiple keys
		keys := []byte{A, B, C} // Typing 'abc'
		for _, key := range keys {
			if err := sendKey(f, key); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
		fmt.Println("Keys sent successfully!")
	})
}

func sendKey(f *os.File, key byte) error {
	// HID report: [Modifier, Reserved, Key1, Key2, Key3, Key4, Key5, Key6]
	press := []byte{KEY_NONE, 0x00, key, 0x00, 0x00, 0x00, 0x00, 0x00}
	release := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// Send key press
	if _, err := f.Write(press); err != nil {
		return fmt.Errorf("failed to send key press: %v", err)
	}

	// Small delay to simulate a real key press
	time.Sleep(50 * time.Millisecond)

	// Send key release
	if _, err := f.Write(release); err != nil {
		return fmt.Errorf("failed to send key release: %v", err)
	}

	return nil
}
