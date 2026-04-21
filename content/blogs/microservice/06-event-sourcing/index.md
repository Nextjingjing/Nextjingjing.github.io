---
title: Microservice ตอนที่ 6 การเก็บประวัติของเหตุการณ์ต่างๆ ด้วย Event sourcing
description: Microservice ตอนที่ 6 การเก็บประวัติของเหตุการณ์ต่างๆ ด้วย Event sourcing
date: 2026-04-18
categories:
    - Blogs
tags:
    - Microservice
    - Backend
toc: true
---

# Event sourcing

> โดยปกติแล้วการบันทึกข้อมูลลงฐานข้อมูลนั้นเราจะบันทึกแค่ค่าปัจจุบันของ Record นั้นๆ แต่ปัญหาที่ตามมาคือเราจะไม่เห็นประวัติเก่าของ Record นั้นๆ ดังนั้น Event sourcing จึงมาแก้ปัญหาดังกล่าว

## ประเภทของการเก็บข้อมูล
1. State-oriented Persistence (การเก็บสถานะปัจจุบัน) เป็นแบบที่เราคุ้นเคยเลยเก็ยสถานะปัจจุบันของข้อมูล แต่ประวัติหายหมด

| account_id | customer_name | balance      | last_updated        | status |
| :--------- | :------------ | :----------- | :------------------ | :----- |
| ACC-001    | Somchai Dee   | **850.00**   | 2024-05-20 10:30:00 | Active |
| ACC-002    | Somsak Range  | **1,200.00** | 2024-05-19 14:15:00 | Active |

2. Event Sourcing (Event-oriented Persistence) เป็นแบบที่เน้นเก็บทุกสถานะของข้อมูลลงในฐานข้อมูล แล้วค่อยนำกลับมารวมกันคำนวณเป็นสถานะปัจจุบันแทน ซึ่งข้อดีคือประวัติยังอยู่

| event_id | account_id | event_type         | amount  | timestamp           |
| :------- | :--------- | :----------------- | :------ | :------------------ |
| 1        | ACC-001    | **AccountOpened**  | 0.00    | 2024-05-20 09:00:00 |
| 2        | ACC-001    | **MoneyDeposited** | 1000.00 | 2024-05-20 09:15:00 |
| 3        | ACC-001    | **MoneyWithdrawn** | 150.00  | 2024-05-20 10:30:00 |
| 4        | ACC-001    | **MoneyWithdrawn** | 200.00  | 2024-05-20 11:00:00 |

ซึ่งสถานะปัจจุบันของ `ACC-001` คือ 0.00 + 1000.00 - 150.00 - 200.00 = **650.00**

## Event sourcing

### ลักษณะ
- Event-oriented Persistence เก็บทุกสถานะแล้วนำมารวมเป็นข้อมูล
- 