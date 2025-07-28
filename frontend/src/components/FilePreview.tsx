import React from 'react';
import { Table, Descriptions, Card } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { CSVData } from '../services/types';

interface FilePreviewProps {
  data: CSVData;
}

const FilePreview: React.FC<FilePreviewProps> = ({ data }) => {
  // 创建表格列
  const columns: ColumnsType<Record<string, string>> = (data.headers || []).map((header, index) => ({
    title: header,
    dataIndex: header,
    key: header,
    width: 150,
    ellipsis: true,
    render: (text) => text || '-',
  }));

  // 转换数据为表格格式
  const tableData = (data.rows || []).map((row, index) => {
    const record: Record<string, string> = { key: index.toString() };
    (data.headers || []).forEach((header, headerIndex) => {
      record[header] = (row || [])[headerIndex] || '';
    });
    return record;
  });

  return (
    <div>
      {/* 文件摘要信息 */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Descriptions column={3} size="small">
          <Descriptions.Item label="总行数">
            {data.summary.total_rows}
          </Descriptions.Item>
          <Descriptions.Item label="总列数">
            {data.summary.total_cols}
          </Descriptions.Item>
          <Descriptions.Item label="预览行数">
            {Math.min((data.rows || []).length, 50)}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 数据表格 */}
      <Table
        columns={columns}
        dataSource={tableData}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => 
            `显示 ${range[0]}-${range[1]} 条，共 ${total} 条数据`,
        }}
        scroll={{ x: 'max-content', y: 400 }}
        size="small"
        bordered
      />
    </div>
  );
};

export default FilePreview;
