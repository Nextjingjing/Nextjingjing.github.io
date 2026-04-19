---
title: Microservice ตอนที่ 5 การเก็บประวัติของเหตุการณ์ต่างๆ ด้วย Event sourcing
description: Microservice ตอนที่ 5 การเก็บประวัติของเหตุการณ์ต่างๆ ด้วย Event sourcing
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

