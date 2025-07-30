import React, { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import { Card, Alert, Spin } from 'antd';

// ECharts配置接口
interface EChartsConfig {
  type: 'bar' | 'line' | 'pie' | 'scatter' | 'heatmap';
  title: string;
  data: Array<{
    name: string;
    value: number | number[];
    [key: string]: any;
  }>;
  xAxis?: string[];
  series?: Array<{
    name: string;
    type: string;
    data: number[];
  }>;
  options?: any;
}

interface EChartsDisplayProps {
  config: EChartsConfig | string;
  width?: string | number;
  height?: string | number;
  loading?: boolean;
}

const EChartsDisplay: React.FC<EChartsDisplayProps> = ({ 
  config, 
  width = '100%', 
  height = 400,
  loading = false
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!chartRef.current || loading) return;

    // 解析配置
    let chartConfig: EChartsConfig;
    try {
      chartConfig = typeof config === 'string' ? JSON.parse(config) : config;
    } catch (error) {
      console.error('ECharts配置解析失败:', error);
      return;
    }

    // 初始化图表
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    // 生成ECharts选项
    const option = generateEChartsOption(chartConfig);
    
    // 设置图表配置
    chartInstance.current.setOption(option, true);

    // 响应式处理
    const handleResize = () => {
      chartInstance.current?.resize();
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [config, loading]);

  useEffect(() => {
    return () => {
      chartInstance.current?.dispose();
    };
  }, []);

  const generateEChartsOption = (config: EChartsConfig) => {
    const baseOption = {
      title: {
        text: config.title,
        left: 'center',
        textStyle: {
          fontSize: 16,
          fontWeight: 'normal'
        }
      },
      tooltip: {
        trigger: 'item'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      }
    };

    switch (config.type) {
      case 'bar':
        return {
          ...baseOption,
          tooltip: {
            trigger: 'axis'
          },
          xAxis: {
            type: 'category',
            data: config.xAxis || config.data.map(item => item.name)
          },
          yAxis: {
            type: 'value'
          },
          series: config.series || [{
            name: '数据',
            type: 'bar',
            data: config.data.map(item => typeof item.value === 'number' ? item.value : item.value[0]),
            itemStyle: {
              color: '#5470c6'
            }
          }]
        };

      case 'line':
        return {
          ...baseOption,
          tooltip: {
            trigger: 'axis'
          },
          xAxis: {
            type: 'category',
            data: config.xAxis || config.data.map(item => item.name)
          },
          yAxis: {
            type: 'value'
          },
          series: config.series || [{
            name: '数据',
            type: 'line',
            data: config.data.map(item => typeof item.value === 'number' ? item.value : item.value[0]),
            lineStyle: {
              color: '#5470c6'
            },
            symbol: 'circle',
            symbolSize: 6
          }]
        };

      case 'pie':
        return {
          ...baseOption,
          tooltip: {
            trigger: 'item',
            formatter: '{a} <br/>{b} : {c} ({d}%)'
          },
          series: [{
            name: config.title,
            type: 'pie',
            radius: '60%',
            center: ['50%', '60%'],
            data: config.data.map(item => ({
              name: item.name,
              value: typeof item.value === 'number' ? item.value : item.value[0]
            })),
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)'
              }
            }
          }]
        };

      case 'scatter':
        return {
          ...baseOption,
          tooltip: {
            trigger: 'item',
            formatter: 'X: {c[0]}<br/>Y: {c[1]}'
          },
          xAxis: {
            type: 'value',
            scale: true
          },
          yAxis: {
            type: 'value',
            scale: true
          },
          series: config.series || [{
            name: '散点数据',
            type: 'scatter',
            data: config.data.map(item => item.value),
            symbolSize: 8,
            itemStyle: {
              color: '#5470c6'
            }
          }]
        };

      case 'heatmap':
        return {
          ...baseOption,
          tooltip: {
            position: 'top',
            formatter: function(params: any) {
              return `${params.marker} ${params.data[2]}`;
            }
          },
          xAxis: {
            type: 'category',
            data: config.xAxis || [],
            splitArea: {
              show: true
            }
          },
          yAxis: {
            type: 'category',
            data: config.options?.yAxis || [],
            splitArea: {
              show: true
            }
          },
          visualMap: {
            min: -1,
            max: 1,
            calculable: true,
            orient: 'horizontal',
            left: 'center',
            bottom: '10%'
          },
          series: [{
            name: '相关性',
            type: 'heatmap',
            data: config.data.map(item => item.value),
            label: {
              show: true
            },
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowColor: 'rgba(0, 0, 0, 0.5)'
              }
            }
          }]
        };

      default:
        return baseOption;
    }
  };

  if (loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Spin size="large" />
          <div style={{ marginTop: '16px' }}>图表生成中...</div>
        </div>
      </Card>
    );
  }

  // 配置解析错误处理
  let isValidConfig = true;
  try {
    const chartConfig = typeof config === 'string' ? JSON.parse(config) : config;
    if (!chartConfig || !chartConfig.data || !Array.isArray(chartConfig.data)) {
      isValidConfig = false;
    }
  } catch (error) {
    isValidConfig = false;
  }

  if (!isValidConfig) {
    return (
      <Alert
        message="图表配置错误"
        description="无法解析ECharts配置，请检查数据格式。"
        type="error"
        showIcon
      />
    );
  }

  return (
    <Card>
      <div 
        ref={chartRef} 
        style={{ 
          width: typeof width === 'number' ? `${width}px` : width, 
          height: typeof height === 'number' ? `${height}px` : height 
        }} 
      />
    </Card>
  );
};

export default EChartsDisplay;
