---
title: "Basic command"
description: "Basic command"
date: 2026-10-13
categories:
    - Blogs
toc: true
---

## Setup lab

ก่อนอื่นเรามา Setup สำหรับการเรียนรู้คำสั่งพื้นกันโดยเราจะใช้ docker กัน
```bash
docker run -d --name redis -p 127.0.0.1:6379:6379 redis:8
```

ทีนี้เราจะมาใช้ `redis-cli` กัน
```bash
docker exec -it redis sh
redis-cli
clear
```

ทีนี้เรามาทำพิธี Hello world กัน รันคำสั่งดังต่อไปนี้
```bash
SET greeting "Hello world"
GET greeting
```
ควรจะเห็นผลลัพธ์ดังต่อไปนี้
```
127.0.0.1:6379> SET greeting "Hello world"
OK
127.0.0.1:6379> GET greeting
"Hello world"
127.0.0.1:6379> 
```

สิ่งที่เราได้ทำตะกี้ก็คือการสร้าง string ที่มีคำว่า "Hello word"

```bash
TYPE greeting # String
```

## คำสั่งพื้นฐาน
เรามาฝึกใช้งานคำสั่งพื้นฐานกัน
### SET
- กำหนดค่าแบบ String ให้กับ Key ครับ
```bash
SET greeting "Hello world"
```
### GET
- อ่านค่าตาม Key
```bash
GET greeting
```

### DEL
- ใช้ลบข้อมูลตาม key
```bash
DEL greeting
```

### MSET
- เป็น SET แต่สร้างหลายๆ ข้อมูลพร้อมกัน
```MSET name "mario" age "25" city "Bangkok"
```

### MGET
- ใช้อ่านข้อมูลหลายๆ Key พร้อมกัน
```bash
MGET name age city
```

เช่น
```
127.0.0.1:6379> MSET name "mario" age "25" city "Bangkok"
OK
127.0.0.1:6379> MGET name age city
1) "mario"
2) "25"
3) "Bangkok"

```

### GETRANGE
- อ่านค่า Key แบบ index start ถึง index end
```bash
GET name # mario
         # 0   4
         # -5 -1
GETRANGE name 0 2 #"mar"
```

### SETRANGE
- เซ็ตค่าตาม Index ของ Key ใดๆ
```bash
SET name "mario"

SETRANGE name 0 "W"

GET name # "Wario"
```

### INCR
```bash
SET views 10
INCR views # (integer) 11
GET views # 11
```

### DECR
```bash
SET views 10
INCR views # (integer) 9
GET views # 9
```

### INCRBY, DECRBY
- เพิ่มเลข โดยใส่จำนวนได้ เช่น
```bash
SET views 10
INCRBY views 10 # (integer) 20
GET views # 20
```

## Options Command
เอาไว้ตั้งค่าคำสั่งพื้นฐานอีกที เช่น `SET session:123 "active" EX 3600 NX` หมายถึงตั้งค่าให้มี Session:123 มีอายุ 3600 วินาที (EX) ส่วน NX รออ่านต่อไปว่าคืออะไร ?

### EX
- ตั้งอายุเป็นวินาที
```
EX seconds
```

### PX
- ตั้งอายุเป็นมิลลิวินาที
```
PX milliseconds
```

### EXAT
- เวลาหมดอายุ ณ เวลา Unix timestamp เช่น `1790000000` ตรงกับ 21 กันยายน 2026 เวลา 21:13:20 น. เวลาไทย
```
EXAT timestamp / PXAT timestamp
```

### NX
- ใช้กับ SET, MSET, ... พวกคำสั่งเขียน
- Only if `Not eXist` งานนี้จะทำตอนที่ key ไม่ปรากฏ
- ไม่เขียนทับ ถ้ามี

### XX
- ใช้กับ SET, MSET, ... พวกคำสั่งเขียน
- Only if `eXists`
- เขียนทับ ถ้ามี

### KEEPTTL
- เขียนทับแต่ใช้ เวลาหมดอายุต่อจากเดิม
```bash
SET session:123 "active" EX 3600
SET session:123 "renewed" KEEPTTL
```

### GET
- คืนค่าก่อนเขียนทับ
```bash
SET name "mario"
SET name "luigi" GET # "mario"
```


{{< page-nav prev-link="/blogs/redis-101/01-redis/" next-link="/post/redis-101/" >}}
