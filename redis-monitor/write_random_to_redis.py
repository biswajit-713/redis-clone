import redis
import uuid
import time
from faker import  Faker

r = redis.Redis(host='localhost', port=7379, db=0)

fake = Faker()

counter = 1
while True:
        random_key = "k" + str(counter)
        random_value = "v" + str(counter)
        r.set(random_key, random_value)
        counter = counter + 1
        time.sleep(0.1)
