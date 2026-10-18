---
title: "Sets, Lists and Hash"
description: "Sets, Lists and Hash"
date: 2026-10-13
categories:
    - Blogs
toc: true
---

## Sets

จากบทที่ 1 เราได้เห็นว่ามันมี Data type ตัวหนึ่งที่เราได้คำอธิบายไปว่า
- สมาชิกใน sets ต้องไม่ซ้ำ
- ไม่มีลำดับ

### การสร้าง

```bash
SADD names mario luigi peach
```

### การอ่าน

```bash
SMEMBERS names
```

### การเพิ่ม
```bash
SADD names yoshi
```

เช่น
```bash
127.0.0.1:6379> SADD names mario luigi peach
(integer) 3
127.0.0.1:6379> SMEMBERS names
1) "mario"
2) "luigi"
3) "peach"
127.0.0.1:6379> SADD names yoshi
(integer) 1
127.0.0.1:6379> SMEMBERS names
1) "mario"
2) "luigi"
3) "peach"
4) "yoshi"
```
#### ข้อควรรู้การเพิ่ม
สมาชิกใน sets ห้ามซ้ำกัน
```bash
127.0.0.1:6379> SMEMBERS names
1) "mario"
2) "luigi"
3) "peach"
4) "yoshi"
127.0.0.1:6379> SADD names yoshi
(integer) 0
127.0.0.1:6379> SMEMBERS names
1) "mario"
2) "luigi"
3) "peach"
4) "yoshi"
```
### การลบ
```bash
SREM names luigi
```

### การรวม 2 sets ด้วย Union
สมมุติมีสองเซ็ต แล้วอยากรวมเป็น 1 ด้วยการ Union
```bash
127.0.0.1:6379> SADD team:a mario luigi
(integer) 2
127.0.0.1:6379> SADD team:b luigi peach
(integer) 2
127.0.0.1:6379> SUNION team:a team:b
1) "peach"
2) "luigi"
3) "mario"
```

### เช็คว่ามีสมาชิกตัวใดๆ ไหม ?
```
SADD team:a mario luigi
SISMEMBER team:a mario # 1 ถ้ามี 0 ถ้าไม่มี
```

### คำสั่งอื่นๆ
- `SINTER team:a team:b` คืนค่าสสมาชิกที่ซ้ำกัน
- `SDIFF team:a team:b` คืนค่าตัวที่มีไม่ซ้ำกัน

### ข้อดี
- สมาชิกไม่ซ้ำกัน เพิ่มสมาชิกซ้ำก็ไม่ทำให้มีตัวซ้ำ
- เช็กสมาชิก เพิ่ม และลบสมาชิกได้รวดเร็ว

### ข้อเสีย
- ไม่มีลำดับ
- ดึงสมาชิกทั้งหมด ได้ช้ากว่า Data ตัวอื่นๆ

### Time complexity
- เพิ่มสมาชิก `SADD` — O(1)
- เช็กสมาชิก `SISMEMBER` — O(1)
- ลบสมาชิก `SREM` — O(N)
- รวม Set `SUNION` — O(N)
- หาจุดร่วม `SINTER` — O(N × M)
- หาส่วนต่าง `SDIFF` — O(N)
- ดึงสมาชิกทั้งหมด `SMEMBERS` — O(N)

## Lists

เป็นกลุ่มของข้อมูลคล้ายๆ Sets แต่จุดต่างคือมีลำดับแน่นอน เช่น Toad, Yoshi, Mario, Peach แต่ให้ดึงมากี่ครั้งก็จะได้ลำดับตามนี้เสมอ

> ไม่ใช่ Array มันคือ Linked lists เข้าถึงข้อมูลตัวแรก และตัวท้ายเร็ว แต่นอกนั้นเข้า นอกนั้น O(n)

### LPUSH
- ใช้สร้าง List 
- เพิ่มสมาชิกตาม Key โดยเป็นการเพิ่มทางซ้าย ของ List

```bash
127.0.0.1:6379> LPUSH students john jane 
(integer) 2
127.0.0.1:6379> LPUSH students peter
(integer) 3
127.0.0.1:6379> LRANGE students 0 -1
1) "peter"
2) "jane"
3) "john"
```

### LRANGE
ใช้เพื่ออ่านข้อมูล เช่น
```bash
127.0.0.1:6379> LRANGE students 0 -1
1) "peter"
2) "jane"
3) "john"
127.0.0.1:6379> LRANGE students 0 1
1) "peter"
2) "jane"
127.0.0.1:6379> LRANGE students 0 999
1) "peter"
2) "jane"
3) "john"
```

### RPUSH
- คล้ายๆ LPUSH แต่เพิ่มทางขวา
```bash
127.0.0.1:6379> RPUSH users john
(integer) 1
127.0.0.1:6379> RPUSH users jane
(integer) 2
127.0.0.1:6379> LRANGE users 0 -1
1) "john"
2) "jane"
```

### LPOP, RPOP
- เอาข้อมูลออกจาก List
  - LPOP คือเอาออกจากทางซ้าย
  - RPOP คือเอาออกจากทางขวา
  - `LPOP key count`
```bash
127.0.0.1:6379> LRANGE students 0 -1
1) "peter"
2) "jane"
3) "john"
127.0.0.1:6379> LPOP students 1
1) "peter"
127.0.0.1:6379> LRANGE students 0 -1
1) "jane"
2) "john"
127.0.0.1:6379> RPOP students 1
1) "john"
127.0.0.1:6379> LRANGE students 0 -1
1) "jane"
```

### LINDEX
- เข้าถึงสมาชิกตัวที่ index
```bash
127.0.0.1:6379> LRANGE students 0 -1
1) "jane"
2) "jane"
3) "john"
4) "jimmy"
127.0.0.1:6379> LINDEX students 2
"john"
```

### LPOS
- หาว่ามีสมาชิกตัวนี้ไหม
- และถ้าหาเจออยู่ที่ index ไหน
  - หากไม่เจอ return nil
```bash
127.0.0.1:6379> LRANGE students 0 -1
1) "jane"
2) "jane"
3) "john"
4) "jimmy"
127.0.0.1:6379> LPOS students jimmy
(integer) 3
127.0.0.1:6379> LPOS students UNKOWN
(nil)
```

### ข้อดี
- เก็บข้อมูลแบบมีลำดับ เพิ่มและนำข้อมูลออกจากหัวหรือท้าย List ได้รวดเร็ว
- เหมาะกับงานแบบคิว (queue) หรือสแตก (stack) เช่น งานที่รอประมวลผล
- เก็บสมาชิกซ้ำกันได้

### ข้อเสีย
- ค้นหาสมาชิกหรือเข้าถึงตำแหน่งกลางๆ ต้องไล่ดูสมาชิก จึงใช้เวลาเพิ่มตามขนาด List
- ถ้าต้องอ่านข้อมูลจำนวนมาก คำสั่งจะใช้เวลาตามจำนวนข้อมูลที่อ่าน
- ไม่เหมาะกับกรณีที่ต้องค้นหาค่าหรือเข้าถึงตำแหน่งใดตำแหน่งหนึ่งบ่อยๆ

### Time complexity
- เพิ่มที่หัว `LPUSH` / เพิ่มที่ท้าย `RPUSH` — O(1) ต่อสมาชิกที่เพิ่ม; ถ้าเพิ่มหลายสมาชิกเป็น O(M) โดย M คือจำนวนสมาชิกที่ส่งเข้าไป
- นำออกจากหัว `LPOP` / นำออกจากท้าย `RPOP` — O(1) ต่อสมาชิกที่นำออก; เมื่อใช้ `count` เป็น O(M)
- อ่านช่วง `LRANGE` — O(S + M), S คือระยะที่ต้องไล่ไปยังจุดเริ่มต้น และ M คือจำนวนสมาชิกที่ส่งคืน
- อ่านตาม index `LINDEX` — O(N) ตามจำนวนสมาชิกที่ต้องไล่ผ่าน; อ่านหัวหรือท้ายเป็น O(1)
- ค้นหาตำแหน่ง `LPOS` — O(N) ในกรณีแย่ที่สุด
- นับจำนวนสมาชิก `LLEN` — O(1)

## Hash
จริงๆ มันก็คือ Hash map ที่เด็กวิศวะคอม/วิทย์คอมได้เรียนมานั้นแหละครับ
```
book:1 -> {
  title -> Name of this wind
  author -> Patrick Rothfuss
  rating -> 5
}
```
> ขอบอกไว้ก่อนนะครับว่า book:1 เป็น key นะครับ อาจจะนิยมใช้เป็น book:***id*** กันครับ

โดยศัพท์ที่เราจะใช้มีดังนี้
- Key หากมองจากตัวอย่างข้างบนมันคือ `book:1`
- Field หากมองตัวอย่างด้านบนได้แก่ `title, author, rating`
- Value หากมองตัวอย่างข้างบนได้แก่ `Name of this wind, Patrick , 5`

### HSET
- ใช้สร้าง Hashmap
- เพิ่มสมาชิก

```bash
127.0.0.1:6379>  HSET book:1 title "The Name of the Wind" author "Patrick Rothfuss" rating "5"
127.0.0.1:6379> HSET book:1 price "500$"
```

```
book:1 -> {
  title -> Name of this wind
  author -> Patrick Rothfuss
  rating -> 5
  price -> "500$"
}
```

### HGET
- ใช้หยิบสมาชิกตาม field
```
book:1 -> {
  title -> Name of this wind
  author -> Patrick Rothfuss
  rating -> 5
  price -> "500$"
}
```

```bash
127.0.0.1:6379> HGET book:1 price
"500$"
```

### HKEYS
- ใช้เพื่อดูว่ามี Fields อะไรบ้าง ?

```
book:1 -> {
  title -> Name of this wind
  author -> Patrick Rothfuss
  rating -> 5
  price -> "500$"
}
```
จะได้ว่า
```bash
127.0.0.1:6379> HKEYS book:1
1) "title"
2) "author"
3) "rating"
4) "price"
```

### HDEL
```
book:1 -> {
  title -> Name of this wind
  author -> Patrick Rothfuss
  rating -> 5
  price -> "500$"
}
```
จะได้ว่า
```bash
127.0.0.1:6379> HDEL book:1 price
(integer) 1
127.0.0.1:6379> HKEYS book:1
1) "title"
2) "author"
3) "rating"
```

### HEXISTS
- ใช้เช็คว่ามี field นี้หรือไม่ ?
```bash
127.0.0.1:6379> HEXISTS book:1 title
(integer) 1 # Exist 
(integer) 0 # Non-exist
```

### ข้อดี
- เก็บข้อมูลที่เกี่ยวข้องกันไว้ใต้ Redis key เดียว เช่น ข้อมูลของหนังสือหนึ่งเล่ม
- อ่านหรือแก้ไขเฉพาะ field ได้ ไม่ต้องเขียนข้อมูลทุก field ใหม่
- เช็กว่ามี field หรือไม่ และอ่านหรือเขียน field เดียวได้รวดเร็ว

### ข้อเสีย
- field ใน Hash ไม่มีลำดับที่รับประกัน
- การอ่าน field และ value ทั้งหมดต้องใช้เวลาตามจำนวน field ใน Hash
- value เป็นข้อมูลแบบ string หากต้องเก็บข้อมูลซ้อนกัน ต้องแปลงเป็นรูปแบบข้อความ เช่น JSON เอง

### Time complexity
ให้ `N` คือจำนวน field/value ที่ระบุในคำสั่งหรือจำนวน field ใน Hash ตามที่อธิบายท้ายแต่ละรายการ
- เพิ่มหรือแก้ไข field `HSET` — O(1) ต่อคู่ field/value; ถ้าส่งหลายคู่เป็น O(N)
- อ่าน field เดียว `HGET` — O(1)
- เช็ก field `HEXISTS` — O(1)
- ลบ field `HDEL` — O(N) เมื่อสั่งลบ N fields
- อ่านชื่อ field ทั้งหมด `HKEYS` — O(N) ตามจำนวน field ใน Hash
- อ่าน field และ value ทั้งหมด `HGETALL` — O(N) ตามจำนวน field ใน Hash

{{< page-nav prev-link="/blogs/redis-101/02-basic-command/" next-link="/post/redis-101/" >}}
