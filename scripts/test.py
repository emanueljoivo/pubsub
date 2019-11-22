import requests
import time
URL = "http://localhost:8003/store"
count = 0

#light message
for i in range(3):
    topic = "topic-"+str(i)
    for j in range(5):
        count+=1
        message = str(j)
        params = {
            "Topic":topic,
            "Message": message,
            "CreatedAt": str(count)
        }
        r = requests.post(url = URL, json=params)

print("light messages sent")
time.sleep(10)
#medium message
for i in range(3):
    topic = "topic-"+str(i)
    for j in range(5):
        count+=1
        message = str(j)*1000
        params = {
            "Topic":topic,
            "Message": message,
            "CreatedAt": str(count)
        }
        r = requests.post(url = URL, json=params)
        
print("medium messages sent")
time.sleep(10)
#heavy message
for i in range(3):
    topic = "topic-"+str(i)
    for j in range(5):
        count+=1
        message = str(j)*100000
        params = {
            "Topic":topic,
            "Message": message,
            "CreatedAt": str(count)
        }
        r = requests.post(url = URL, json=params)
print("heavy messages sent")