import React from 'react';
import { Avatar, Card } from 'antd';
import { RobotOutlined, UserOutlined } from '@ant-design/icons';
import MarkdownRenderer from './MarkdownRenderer';

interface ChatMessageProps {
  type: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

const ChatMessage: React.FC<ChatMessageProps> = ({ type, content, timestamp }) => {
  const isUser = type === 'user';
  
  // 检测内容是否包含ECharts配置
  const hasEChartsConfig = content.includes('```json') && 
                           (content.includes('"type"') && content.includes('"data"')) ||
                           content.includes('```echarts') ||
                           content.includes('ECHARTS_CONFIG_START');
  
  return (
    <div style={{ 
      display: 'flex', 
      marginBottom: 12,
      flexDirection: isUser ? 'row-reverse' : 'row'
    }}>
      <Avatar 
        icon={isUser ? <UserOutlined /> : <RobotOutlined />}
        style={{ 
          backgroundColor: isUser ? '#1890ff' : '#52c41a',
          flexShrink: 0,
          margin: isUser ? '0 0 0 8px' : '0 8px 0 0'
        }}
      />
      <div style={{ 
        maxWidth: isUser ? '70%' : '85%', // AI消息可以更宽，支持图表显示
        display: 'flex',
        flexDirection: 'column',
        alignItems: isUser ? 'flex-end' : 'flex-start'
      }}>
        <Card
          size="small"
          style={{
            backgroundColor: isUser ? '#e6f7ff' : '#f6ffed',
            border: `1px solid ${isUser ? '#91d5ff' : '#b7eb8f'}`,
            borderRadius: 8,
            marginBottom: 4,
            width: '100%'
          }}
          bodyStyle={{ padding: 12 }}
        >
          {!isUser && hasEChartsConfig ? (
            // 使用Markdown渲染器处理包含ECharts的内容
            <MarkdownRenderer content={content} />
          ) : (
            <div style={{ 
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              fontSize: 14,
              lineHeight: 1.6
            }}>
              {content}
            </div>
          )}
        </Card>
        <div style={{ 
          fontSize: 11, 
          color: '#999',
          padding: '0 4px',
        }}>
          {new Date(timestamp).toLocaleTimeString()}
        </div>
      </div>
    </div>
  );
};

export default ChatMessage;
