import React from 'react';
import { Card, Row, Col, Statistic } from 'antd';
import { Line, Column, Pie } from '@ant-design/plots';
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';

interface StatsChartsProps {
  stats: any;
}

export const StatsCharts: React.FC<StatsChartsProps> = ({ stats }) => {
  if (!stats) return null;

  // Prepare data for hourly chart
  const hourlyData = stats.hourly?.map((item: any) => ({
    hour: item.hour,
    value: item.visits,
    type: 'Visits',
  })) || [];

  // Prepare data for geographic distribution
  const geoData = stats.geographic?.slice(0, 10) || [];

  // Prepare data for target distribution  
  const targetData = stats.top_targets?.map((item: any) => ({
    target: `Target ${item.target_id}`,
    hits: item.hits,
    percentage: item.percentage,
  })) || [];

  const hourlyConfig = {
    data: hourlyData,
    xField: 'hour',
    yField: 'value',
    seriesField: 'type',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    xAxis: {
      label: {
        autoRotate: true,
        autoHide: true,
      },
    },
    yAxis: {
      label: {
        formatter: (v: string) => `${v}`,
      },
    },
    tooltip: {
      formatter: (datum: any) => {
        return { name: 'Visits', value: datum.value };
      },
    },
  };

  const geoConfig = {
    data: geoData,
    xField: 'country_name',
    yField: 'count',
    color: '#5B8FF9',
    label: {
      position: 'top',
      style: {
        fill: '#FFFFFF',
        opacity: 0.6,
      },
    },
    xAxis: {
      label: {
        autoHide: true,
        autoRotate: false,
      },
    },
    meta: {
      country_name: {
        alias: 'Country',
      },
      count: {
        alias: 'Visits',
      },
    },
  };

  const targetConfig = {
    data: targetData,
    angleField: 'hits',
    colorField: 'target',
    radius: 0.8,
    label: {
      type: 'outer',
      content: '{name} {percentage}%',
    },
    interactions: [
      {
        type: 'pie-legend-active',
      },
      {
        type: 'element-active',
      },
    ],
  };

  return (
    <div className="stats-charts">
      <Row gutter={[16, 16]}>
        {/* Summary Cards */}
        <Col span={6}>
          <Card>
            <Statistic
              title="Total Links"
              value={stats.summary?.total_links || 0}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Today's Visits"
              value={stats.summary?.today_visits || 0}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="This Week"
              value={stats.summary?.week_visits || 0}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Success Rate"
              value={stats.summary?.success_rate || '0%'}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        {/* Hourly Traffic Chart */}
        <Col span={24}>
          <Card title="Hourly Traffic">
            <Line {...hourlyConfig} height={300} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        {/* Geographic Distribution */}
        <Col span={12}>
          <Card title="Geographic Distribution">
            <Column {...geoConfig} height={300} />
          </Card>
        </Col>

        {/* Target Distribution */}
        <Col span={12}>
          <Card title="Target Distribution">
            <Pie {...targetConfig} height={300} />
          </Card>
        </Col>
      </Row>
    </div>
  );
};