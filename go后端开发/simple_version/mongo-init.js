// MongoDB初始化脚本
// 用于创建YAML配置管理器所需的数据库和集合

// 切换到指定的数据库
db = db.getSiblingDB('vnf_config');

// 创建用户（如果需要）
db.createUser({
  user: 'app_user',
  pwd: 'app_password',
  roles: [
    {
      role: 'readWrite',
      db: 'vnf_config'
    }
  ]
});

// 创建集合和索引
db.createCollection('yaml_latest');
db.createCollection('yaml_history');

// 为yaml_latest集合创建索引
db.yaml_latest.createIndex({ "filename": 1 }, { unique: true });
db.yaml_latest.createIndex({ "updated_at": -1 });

// 为yaml_history集合创建索引
db.yaml_history.createIndex({ "filename": 1 });
db.yaml_history.createIndex({ "timestamp": -1 });
db.yaml_history.createIndex({ "filename": 1, "timestamp": -1 });

print('MongoDB初始化完成：数据库 vnf_config 已创建，集合和索引已设置');








