import time 

from button import *

b = Button(25, debounce=0.1)

print("checking")
while True:
    if b.is_pressed():
        print(time.time())
    time.sleep(500) # Sleep for 500 milliseconds
    print("polling")
