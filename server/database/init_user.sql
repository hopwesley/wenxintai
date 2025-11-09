-- 初始化数据库用户 wenxintai，并赋予 hyperorchid 库访问权限
-- 请在 psql 中以 postgres 用户执行： \i init_user.sql

-- 1️⃣ 创建用户 wenxintai，设置密码
CREATE USER wenxintai WITH PASSWORD 'wenxintai_password';

-- 2️⃣ 授权 wenxintai 连接数据库 hyperorchid
GRANT CONNECT ON DATABASE hyperorchid TO wenxintai;

-- 3️⃣ 切换到目标数据库
\c hyperorchid

-- 4️⃣ 授权 schema 与表访问
GRANT USAGE ON SCHEMA public TO wenxintai;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO wenxintai;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO wenxintai;

-- 5️⃣ 确保未来新建表/序列也自动授权
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO wenxintai;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO wenxintai;

-- 6️⃣ （可选）查看结果
\du
