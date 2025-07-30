import React from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { tomorrow } from 'react-syntax-highlighter/dist/esm/styles/prism';
import EChartsDisplay from './EChartsDisplay';

interface MarkdownRendererProps {
  content: string;
}

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ content }) => {
  return (
    <ReactMarkdown
      components={{
        code({ node, inline, className, children, ...props }: any) {
          const match = /language-(\w+)/.exec(className || '');
          const language = match ? match[1] : '';

          if (!inline && language === 'echarts') {
            // 渲染ECharts图表
            try {
              const chartConfig = String(children).replace(/\n$/, '');
              return <EChartsDisplay config={chartConfig} />;
            } catch (error) {
              console.error('ECharts渲染错误:', error);
              return (
                <div style={{ 
                  padding: '16px', 
                  background: '#fff2f0', 
                  border: '1px solid #ffccc7',
                  borderRadius: '6px',
                  color: '#a8071a'
                }}>
                  ECharts图表渲染失败: {String(error)}
                </div>
              );
            }
          } else if (!inline && language === 'json' && String(children).includes('"type"') && String(children).includes('"data"')) {
            // 尝试检测是否为ECharts配置的JSON
            try {
              const jsonContent = String(children).replace(/\n$/, '');
              const config = JSON.parse(jsonContent);
              if (config.type && config.data && Array.isArray(config.data)) {
                return <EChartsDisplay config={config} />;
              }
            } catch (error) {
              // 如果不是ECharts配置，fallback到代码高亮
            }
          }

          if (!inline && match) {
            return (
              <SyntaxHighlighter
                style={tomorrow}
                language={language}
                PreTag="div"
                {...props}
              >
                {String(children).replace(/\n$/, '')}
              </SyntaxHighlighter>
            );
          }

          return (
            <code className={className} {...props}>
              {children}
            </code>
          );
        },
        // 自定义图片渲染，支持ECharts配置
        img({ src, alt, ...props }: any) {
          if (alt && alt.toLowerCase().includes('echarts')) {
            try {
              // 如果图片alt包含echarts，并且src是JSON配置
              if (src && src.startsWith('data:application/json')) {
                const jsonData = atob(src.split(',')[1]);
                return <EChartsDisplay config={jsonData} />;
              }
            } catch (error) {
              console.error('ECharts图片渲染错误:', error);
            }
          }
          return <img src={src} alt={alt} {...props} />;
        }
      }}
    >
      {content}
    </ReactMarkdown>
  );
};

export default MarkdownRenderer;
