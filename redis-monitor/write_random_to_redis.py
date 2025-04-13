import redis
import uuid
import time
from faker import  Faker

r = redis.Redis(host='localhost', port=7379, db=0)

fake = Faker()

while True:
    for i in range(0, 58):
        random_key = f"key:{uuid.uuid4()}"
        random_value = fake.sentence()

        r.set(random_key, random_value)
    time.sleep(20)
