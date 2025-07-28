import React from 'react';
import { Avatar, Card } from 'antd';
import { RobotOutlined, UserOutlined } from '@ant-design/icons';

interface ChatMessageProps {
  type: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

const ChatMessage: React.FC<ChatMessageProps> = ({ type, content, timestamp }) => {
  const isUser = type === 'user';
  
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
        maxWidth: '70%',
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
          }}
          bodyStyle={{ padding: 12 }}
        >
          <div style={{ 
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontSize: 14,
            lineHeight: 1.6
          }}>
            {content}
          </div>
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
