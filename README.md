# Golang Task - Social Media API

این پروژه یک API ساده است که به کاربران اجازه می‌دهد پست‌هایی ایجاد، ویرایش و حذف کنند و همچنین تایم‌لاین خود را مشاهده کنند. پروژه از Redis برای کش کردن و MySQL برای ذخیره‌سازی داده‌ها استفاده می‌کند.

## پیش‌نیازها

قبل از شروع، لطفاً اطمئان حاصل کنید که این ابزارها روی سیستم شما نصب شده‌اند:

<ol>
    <li><a href="https://golang.org/doc/install">Go</a> (نسخه 1.25 یا بالاتر)</li>
    <li><a href="https://dev.mysql.com/downloads/installer/">MySQL</a></li>
    <li><a href="https://redis.io/download">Redis</a></li>
</ol>

## نصب پروژه

### <h2>کلون کردن پروژه :</h2>

ابتدا پروژه را از گیت‌هاب کلون کنید:

```bash
git clone https://github.com/MahdiMalvandi/golang_task.git
cd golang_task
```
<h2>نصب وابستگی‌ها:</h2>

با استفاده از Go وابستگی‌های پروژه را نصب کنید:

```bash
go mod download
```
<h2>ایجاد دیتابیس MySQL:</h2>

قبل از راه‌اندازی پروژه، باید دیتابیس MySQL را راه‌اندازی کنید. اگر به طور محلی از MySQL استفاده می‌کنید، می‌توانید به سیستم MySQL خود متصل شوید و دیتابیس را بسازید:
```bash
CREATE DATABASE your_db_name;
```
<h3>ایجاد دیتابیس با Docker:</h3>

اگر از Docker برای راه‌اندازی MySQL استفاده می‌کنید، می‌توانید از دستورات زیر استفاده کنید:
```bash
docker run --name mysql_db -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=mydb -p 3306:3306 -d mysql:8.0
docker exec -it mysql_db mysql -uroot -p
CREATE DATABASE new_database_name;
```

در اینجا:

<ul> <li><code>MYSQL_ROOT_PASSWORD=1234</code>: رمز عبور برای کاربر <code>root</code>.</li> <li><code>MYSQL_DATABASE=mydb</code>: نام پایگاه داده پیش‌فرض.</li> <li><code>3306:3306</code>: اتصال پورت 3306 کانتینر به سیستم میزبان.</li> <li><code>new_database_name</code>: نام دیتابی که می‌خواهید بسازید.</li> </ul>

اگر از Docker استفاده می‌کنید، در فایل <code>.env</code> به جای <code>localhost</code>، باید اطلاعات کانتینر Docker خود را وارد کنید:


<ul> <li><code>DB_HOST=mysql_db</code> (نام کانتینر MySQL شما)</li> <li><code>DB_USER=root</code></li> <li><code>DB_PASSWORD=1234</code></li> <li><code>DB_NAME=new_database_name</code> (نام دیتابی که ساختید)</li> </ul>
<h2>راه‌اندازی Redis:</h2>

برای راه‌اندازی Redis با Docker، می‌توانید از دستور زیر استفاده کنید:
```bash
docker run --name redis -p 6379:6379 -d redis
```

این دستور یک کانتینر Redis ایجاد می‌کند که از پورت 6379 در سیستم شما در دسترس خواهد بود.

<h2>پیکربندی محیط:</h2>

پس از ایجاد دیتابیس MySQL و راه‌اندازی Redis، شما باید یک فایل <code>.env</code> در پوشه ریشه پروژه ایجاد کنید و اطلاعات اتصال به دیتابیس و Redis را در آن قرار دهید. این فایل باید به شکل زیر باشد:

```bash
DB_HOST=localhost
DB_USER=root
DB_PASSWORD=1234
DB_NAME=mydb
REDIS_ADDR=localhost:6379
JWT_SECRET=your_secret_key
MAX_FILE_SIZE=50
```

توجه:

<ul> <li>در صورتی که از Docker استفاده می‌کنید، مقدار <code>DB_HOST</code> باید نام کانتینر MySQL (<code>mysql_db</code>) باشد.</li> <li>برای <code>REDIS_ADDR</code> هم باید از همان پورت 6379 استفاده کنید.</li> <li><code>JWT_SECRET</code> باید یک کلید محرمانه تصادفی و پیچیده باشد.</li> </ul>
<h2>راه‌اندازی API و تست آن:</h2>

پس از پیکربندی محیط، برای راه‌اندازی سرور Go API، دستور زیر را اجرا کنید:
```bash
go run main.go
```

سپس می‌توانید از طریق مرورگر خود به مستندات Swagger API که در مسیر <code>/swagger/index.html</code> در دسترس است، دسترسی پیدا کنید.

<h2>دسترسی به مستندات API:</h2>

پس از راه‌اندازی سرور، می‌توانید به مستندات Swagger API از طریق مسیر زیر دسترسی پیدا کنید:
```bash
http://localhost:3001/swagger/index.html
```

در این صفحه، شما می‌توانید تمامی عملیات‌ها و متدهای موجود API را مشاهده کنید.