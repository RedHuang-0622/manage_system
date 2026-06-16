-- Lab Management System — Seed Data
-- Run: mysql -u root -p lab_manage < seed.sql
-- Encoding: UTF-8

SET NAMES utf8mb4;

-- Roles (幂等)
INSERT IGNORE INTO sys_roles (role_name, description, is_system, created_at, updated_at) VALUES
('super_admin',       '超级管理员（指导老师）', 1, NOW(), NOW()),
('lab_admin',         '实验室负责人',           1, NOW(), NOW()),
('equipment_manager', '设备管理员',             1, NOW(), NOW()),
('member',            '普通成员',               1, NOW(), NOW()),
('viewer',            '观察员（只读）',         1, NOW(), NOW());

-- Super admin (password: admin123, bcrypt cost 12)
INSERT IGNORE INTO sys_users (username, password_hash, real_name, role_id, status, created_at, updated_at)
SELECT 'admin', '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '系统管理员',
       (SELECT id FROM sys_roles WHERE role_name='super_admin'), 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sys_users WHERE username='admin');

-- Demo users (password: 123456)
INSERT IGNORE INTO sys_users (username, password_hash, real_name, email, phone, role_id, status, created_at, updated_at) VALUES
('liulei',    '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '刘磊',  'liulei@lab.edu.cn',    '13800138006', (SELECT id FROM sys_roles WHERE role_name='equipment_manager'), 1, NOW(), NOW()),
('ling',      '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '凌',    'ling@lab.edu.cn',      '13800138007', (SELECT id FROM sys_roles WHERE role_name='member'),            1, NOW(), NOW()),
('red',       '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', 'Red',   'red@lab.edu.cn',       '13800138008', (SELECT id FROM sys_roles WHERE role_name='member'),            1, NOW(), NOW()),
('zhangwei',  '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '张伟',  'zhangwei@lab.edu.cn',  '13800138001', (SELECT id FROM sys_roles WHERE role_name='lab_admin'),        1, NOW(), NOW()),
('lina',      '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '李娜',  'lina@lab.edu.cn',      '13800138002', (SELECT id FROM sys_roles WHERE role_name='member'),            1, NOW(), NOW()),
('wangqiang', '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '王强',  'wangqiang@lab.edu.cn', '13800138003', (SELECT id FROM sys_roles WHERE role_name='member'),            1, NOW(), NOW()),
('zhaomin',   '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '赵敏',  'zhaomin@lab.edu.cn',   '13800138004', (SELECT id FROM sys_roles WHERE role_name='member'),            1, NOW(), NOW()),
('chenjing',  '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '陈静',  'chenjing@lab.edu.cn',  '13800138005', (SELECT id FROM sys_roles WHERE role_name='member'),            1, NOW(), NOW()),
('sunyue',    '$2a$12$I0tjXEMTuaCq4fLYP4MjvuZTAl8hI1Uw5PJnk6PC8NgWvIKqQGFBq', '孙悦',  'sunyue@lab.edu.cn',    '13800138009', (SELECT id FROM sys_roles WHERE role_name='viewer'),            1, NOW(), NOW());

-- Equipment
INSERT IGNORE INTO lab_equipments (name, model, category, total_stock, available_stock, location, status, created_at, updated_at) VALUES
('示波器 TDS2024C',        'TDS2024C',             '测量仪器',   3,  3,  'A301实验室',   1, NOW(), NOW()),
('GPU服务器 DGX-A100',     'DGX-A100',             '服务器',     4,  4,  'A302机房',     1, NOW(), NOW()),
('3D打印机 Ultimaker S5',  'Ultimaker S5',         '制造设备',   2,  2,  'B101创客空间', 1, NOW(), NOW()),
('频谱分析仪 N9020B',      'Keysight N9020B',      '测量仪器',   1,  1,  'A303实验室',   1, NOW(), NOW()),
('MacBook Pro M3',         'MacBook Pro 16 M3 Pro','笔记本电脑', 10, 10, 'B202设备室',   1, NOW(), NOW()),
('逻辑分析仪 16862A',      'Keysight 16862A',      '测量仪器',   2,  2,  'A304实验室',   1, NOW(), NOW());
