import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Upload,
  message,
  Space,
  Tag,
  Modal,
  Tooltip,
} from 'antd';
import {
  UploadOutlined,
  DeleteOutlined,
  EyeOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UploadProps } from 'antd';
import { fileService } from '../services/file';
import { FileInfo, CSVData } from '../services/types';
import FilePreview from '../components/FilePreview';

const { confirm } = Modal;

const Files: React.FC = () => {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [previewData, setPreviewData] = useState<CSVData | null>(null);
  const [selectedFile, setSelectedFile] = useState<FileInfo | null>(null);

  useEffect(() => {
    loadFiles();
  }, []);

  const loadFiles = async () => {
    setLoading(true);
    try {
      const data = await fileService.getList();
      setFiles(data);
    } catch (error: any) {
      message.error('加载文件列表失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options;
    setUploading(true);

    try {
      const uploadedFile = await fileService.upload(file as File);
      message.success('文件上传成功！');
      loadFiles(); // 重新加载文件列表
      onSuccess?.(uploadedFile);
    } catch (error: any) {
      message.error('文件上传失败: ' + error.message);
      onError?.(error);
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = (file: FileInfo) => {
    confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除文件 "${file.orig_name}" 吗？此操作不可恢复。`,
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await fileService.delete(file.id);
          message.success('文件删除成功！');
          loadFiles();
        } catch (error: any) {
          message.error('文件删除失败: ' + error.message);
        }
      },
    });
  };

  const handlePreview = async (file: FileInfo) => {
    if (file.status !== 'ready') {
      message.warning('文件还未处理完成，无法预览');
      return;
    }

    try {
      const data = await fileService.preview(file.id, 50);
      setPreviewData(data);
      setSelectedFile(file);
      setPreviewVisible(true);
    } catch (error: any) {
      message.error('预览文件失败: ' + error.message);
    }
  };

  const getStatusTag = (status: string) => {
    const statusMap = {
      uploaded: { color: 'processing', text: '已上传' },
      processing: { color: 'warning', text: '处理中' },
      ready: { color: 'success', text: '就绪' },
      error: { color: 'error', text: '错误' },
    };
    const config = statusMap[status as keyof typeof statusMap] || { color: 'default', text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const columns: ColumnsType<FileInfo> = [
    {
      title: '文件名',
      dataIndex: 'orig_name',
      key: 'orig_name',
      render: (text, record) => (
        <Space>
          <FileTextOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '文件类型',
      dataIndex: 'type',
      key: 'type',
      render: (type) => <Tag>{type.toUpperCase()}</Tag>,
    },
    {
      title: '文件大小',
      dataIndex: 'size',
      key: 'size',
      render: (size) => formatFileSize(size),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: getStatusTag,
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="预览">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handlePreview(record)}
              disabled={record.status !== 'ready'}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    customRequest: handleUpload,
    showUploadList: false,
    accept: '.csv,.xlsx,.xls,.json',
    beforeUpload: (file) => {
      const isValidType = ['csv', 'xlsx', 'xls', 'json'].some(ext => 
        file.name.toLowerCase().endsWith(`.${ext}`)
      );
      if (!isValidType) {
        message.error('只支持上传 CSV、Excel 和 JSON 文件！');
        return false;
      }
      const isLt500M = file.size / 1024 / 1024 < 500;
      if (!isLt500M) {
        message.error('文件大小不能超过 500MB！');
        return false;
      }
      return true;
    },
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h1>文件管理</h1>
        <Upload {...uploadProps}>
          <Button 
            type="primary" 
            icon={<UploadOutlined />}
            loading={uploading}
          >
            上传文件
          </Button>
        </Upload>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={files}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个文件`,
          }}
        />
      </Card>

      <Modal
        title={`预览文件: ${selectedFile?.orig_name}`}
        open={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        width={1000}
        footer={null}
      >
        {previewData && <FilePreview data={previewData} />}
      </Modal>
    </div>
  );
};

export default Files;
