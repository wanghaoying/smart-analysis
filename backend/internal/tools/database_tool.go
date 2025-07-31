package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// DatabaseTool 数据库连接工具
type DatabaseTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewDatabaseTool 创建数据库连接工具
func NewDatabaseTool(sandbox *sanbox.PythonSandbox) *DatabaseTool {
	return &DatabaseTool{
		sandbox: sandbox,
		name:    "database_tool",
		desc:    "数据库连接和查询工具，支持MySQL、PostgreSQL、SQLite等常见数据库的连接和数据提取。",
	}
}

// Info 返回工具信息
func (t *DatabaseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"db_type": {
					Type:     schema.String,
					Desc:     "数据库类型: mysql, postgresql, sqlite, oracle",
					Required: true,
				},
				"connection_string": {
					Type:     schema.String,
					Desc:     "数据库连接字符串",
					Required: true,
				},
				"query": {
					Type:     schema.String,
					Desc:     "SQL查询语句",
					Required: true,
				},
				"limit": {
					Type:     schema.Number,
					Desc:     "查询结果限制行数（默认1000）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *DatabaseTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		DbType           string `json:"db_type"`
		ConnectionString string `json:"connection_string"`
		Query            string `json:"query"`
		Limit            int    `json:"limit,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	if args.Limit <= 0 {
		args.Limit = 1000
	}

	code := t.generateDatabaseCode(args.DbType, args.ConnectionString, args.Query, args.Limit)

	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "数据库查询失败: " + result.Error, nil
	}

	return result.Stdout, nil
}

// generateDatabaseCode 生成数据库连接代码
func (t *DatabaseTool) generateDatabaseCode(dbType, connectionString, query string, limit int) string {
	code := fmt.Sprintf(`
import pandas as pd
import json

# 数据库连接配置
db_type = "%s"
connection_string = "%s"
query = """
%s
"""
limit = %d

try:
    # 根据数据库类型选择连接方式
    if db_type.lower() == "sqlite":
        import sqlite3
        conn = sqlite3.connect(connection_string)
        df = pd.read_sql_query(query, conn)
        conn.close()
        
    elif db_type.lower() == "mysql":
        try:
            import pymysql
            # 解析连接字符串（简化版本）
            if connection_string.startswith('mysql://'):
                import sqlalchemy
                engine = sqlalchemy.create_engine(connection_string)
                df = pd.read_sql_query(query, engine)
                engine.dispose()
            else:
                print(json.dumps({"error": "MySQL连接需要完整的连接字符串"}, ensure_ascii=False))
                exit()
        except ImportError:
            print(json.dumps({"error": "缺少pymysql或sqlalchemy库"}, ensure_ascii=False))
            exit()
            
    elif db_type.lower() == "postgresql":
        try:
            import psycopg2
            import sqlalchemy
            if connection_string.startswith('postgresql://'):
                engine = sqlalchemy.create_engine(connection_string)
                df = pd.read_sql_query(query, engine)
                engine.dispose()
            else:
                print(json.dumps({"error": "PostgreSQL连接需要完整的连接字符串"}, ensure_ascii=False))
                exit()
        except ImportError:
            print(json.dumps({"error": "缺少psycopg2或sqlalchemy库"}, ensure_ascii=False))
            exit()
            
    else:
        print(json.dumps({"error": f"不支持的数据库类型: {db_type}"}, ensure_ascii=False))
        exit()
    
    # 限制结果行数
    if len(df) > limit:
        df = df.head(limit)
        truncated = True
    else:
        truncated = False
    
    # 格式化结果
    result = {
        "success": True,
        "db_type": db_type,
        "query": query,
        "data": {
            "shape": df.shape,
            "columns": df.columns.tolist(),
            "dtypes": df.dtypes.astype(str).to_dict(),
            "records": df.to_dict("records"),
            "truncated": truncated,
            "total_rows": len(df)
        }
    }
    
    print(json.dumps(result, ensure_ascii=False, indent=2, default=str))
    
except Exception as e:
    error_result = {
        "success": False,
        "error": str(e),
        "db_type": db_type
    }
    print(json.dumps(error_result, ensure_ascii=False, indent=2))

`, dbType, connectionString, query, limit)

	return code
}
