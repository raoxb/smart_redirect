import React, { useState } from 'react'
import { Card, Row, Col, Statistic, Select, DatePicker, Table, Tag } from 'antd'
import { Line, Column, Pie } from '@ant-design/plots'
import { useQuery } from '@tanstack/react-query'
import { BarChartOutlined, LinkOutlined, GlobalOutlined, EyeOutlined } from '@ant-design/icons'
import { api } from '@/services/api'
import type { ColumnsType } from 'antd/es/table'

const { RangePicker } = DatePicker
const { Option } = Select

interface LinkStats {
  link_id: string
  business_unit: string
  total_hits: number
  today_hits: number
  unique_ips: number
  countries: { country: string; hits: number }[]
  targets: { target_id: number; url: string; hits: number }[]
}

interface CountryStats {
  country: string
  hits: number
}

interface SystemStats {
  total_links: number
  total_hits: number
  today_hits: number
  unique_ips: number
  top_countries: CountryStats[]
}

const Statistics: React.FC = () => {
  const [selectedPeriod, setSelectedPeriod] = useState('7d')
  const [selectedLink, setSelectedLink] = useState<string | null>(null)

  // Fetch system statistics
  const { data: systemStats, isLoading: systemLoading } = useQuery<SystemStats>({
    queryKey: ['systemStats'],
    queryFn: () => api.get('/stats/system').then(res => res.data),
  })

  // Fetch all links for selection
  const { data: linksData } = useQuery({
    queryKey: ['links'],
    queryFn: () => api.get('/links').then(res => res.data),
  })

  // Fetch specific link statistics
  const { data: linkStats, isLoading: linkLoading } = useQuery<LinkStats>({
    queryKey: ['linkStats', selectedLink],
    queryFn: () => api.get(`/stats/links/${selectedLink}`).then(res => res.data),
    enabled: !!selectedLink,
  })

  // Fetch hourly data for selected link
  const { data: hourlyData } = useQuery({
    queryKey: ['hourlyStats', selectedLink, selectedPeriod],
    queryFn: () => {
      const hours = selectedPeriod === '24h' ? 24 : selectedPeriod === '7d' ? 168 : 24
      return api.get(`/stats/links/${selectedLink}/hourly?hours=${hours}`).then(res => res.data)
    },
    enabled: !!selectedLink,
  })

  // Top links table columns
  const topLinksColumns: ColumnsType<any> = [
    {
      title: 'Link ID',
      dataIndex: 'link_id',
      key: 'link_id',
      render: (linkId: string) => (
        <Tag color="blue" style={{ cursor: 'pointer' }} onClick={() => setSelectedLink(linkId)}>
          {linkId}
        </Tag>
      ),
    },
    {
      title: 'Business Unit',
      dataIndex: 'business_unit',
      key: 'business_unit',
    },
    {
      title: 'Total Hits',
      dataIndex: 'current_hits',
      key: 'current_hits',
      sorter: (a, b) => a.current_hits - b.current_hits,
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
  ]

  // Prepare chart data
  const countryChartData = systemStats?.top_countries?.map(item => ({
    country: item.country || 'Unknown',
    hits: item.hits,
  })) || []

  const targetChartData = linkStats?.targets?.map(item => ({
    target: item.url.split('/').pop() || 'Unknown',
    hits: item.hits,
  })) || []

  const hourlyChartData = hourlyData?.map((item: any) => ({
    hour: new Date(item.hour).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
    hits: item.hits,
  })) || []

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <h1>Statistics Dashboard</h1>
        <p>Comprehensive analytics and performance metrics</p>
      </div>

      {/* System Overview Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Total Links"
              value={systemStats?.total_links || 0}
              prefix={<LinkOutlined />}
              loading={systemLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Total Hits"
              value={systemStats?.total_hits || 0}
              prefix={<EyeOutlined />}
              loading={systemLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Today's Hits"
              value={systemStats?.today_hits || 0}
              prefix={<BarChartOutlined />}
              loading={systemLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Unique IPs"
              value={systemStats?.unique_ips || 0}
              prefix={<GlobalOutlined />}
              loading={systemLoading}
            />
          </Card>
        </Col>
      </Row>

      {/* Filter Controls */}
      <Card style={{ marginBottom: '24px' }}>
        <Row gutter={16} align="middle">
          <Col>
            <label style={{ marginRight: '8px' }}>Select Link:</label>
            <Select
              style={{ width: 200 }}
              placeholder="Choose a link"
              value={selectedLink}
              onChange={setSelectedLink}
              allowClear
            >
              {linksData?.data?.map((link: any) => (
                <Option key={link.link_id} value={link.link_id}>
                  {link.link_id} - {link.business_unit}
                </Option>
              ))}
            </Select>
          </Col>
          <Col>
            <label style={{ marginRight: '8px' }}>Time Period:</label>
            <Select
              style={{ width: 120 }}
              value={selectedPeriod}
              onChange={setSelectedPeriod}
            >
              <Option value="24h">24 Hours</Option>
              <Option value="7d">7 Days</Option>
            </Select>
          </Col>
        </Row>
      </Card>

      {/* Charts Section */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        {/* Hourly Traffic Chart */}
        <Col xs={24} lg={12}>
          <Card title="Hourly Traffic" loading={linkLoading}>
            {hourlyChartData.length > 0 ? (
              <Line
                data={hourlyChartData}
                xField="hour"
                yField="hits"
                height={300}
                smooth={true}
                point={{ size: 4 }}
                tooltip={{
                  formatter: (datum) => ({
                    name: 'Hits',
                    value: datum.hits,
                  }),
                }}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: '50px', color: '#999' }}>
                {selectedLink ? 'No data available for selected period' : 'Select a link to view hourly traffic'}
              </div>
            )}
          </Card>
        </Col>

        {/* Geographic Distribution */}
        <Col xs={24} lg={12}>
          <Card title="Geographic Distribution">
            {countryChartData.length > 0 ? (
              <Column
                data={countryChartData}
                xField="country"
                yField="hits"
                height={300}
                label={{
                  position: 'top',
                  formatter: (text) => `${text}`,
                }}
                tooltip={{
                  formatter: (datum) => ({
                    name: 'Country',
                    value: `${datum.country}: ${datum.hits} hits`,
                  }),
                }}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: '50px', color: '#999' }}>
                No geographic data available
              </div>
            )}
          </Card>
        </Col>

        {/* Target Distribution */}
        <Col xs={24} lg={12}>
          <Card title="Target Distribution" loading={linkLoading}>
            {targetChartData.length > 0 ? (
              <Pie
                data={targetChartData}
                angleField="hits"
                colorField="target"
                height={300}
                radius={0.8}
                label={{
                  type: 'outer',
                  content: (data) => `${data.target}: ${((data.hits / targetChartData.reduce((sum, item) => sum + item.hits, 0)) * 100).toFixed(1)}%`,
                }}
                tooltip={{
                  formatter: (datum) => ({
                    name: datum.target,
                    value: `${datum.hits} hits`,
                  }),
                }}
              />
            ) : (
              <div style={{ textAlign: 'center', padding: '50px', color: '#999' }}>
                {selectedLink ? 'No target data available' : 'Select a link to view target distribution'}
              </div>
            )}
          </Card>
        </Col>

        {/* Top Links Table */}
        <Col xs={24} lg={12}>
          <Card title="Top Links">
            <Table
              columns={topLinksColumns}
              dataSource={linksData?.data || []}
              rowKey="id"
              pagination={{ pageSize: 5 }}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      {/* Link Details Section */}
      {selectedLink && linkStats && (
        <Card title={`Link Details: ${selectedLink}`} style={{ marginBottom: '24px' }}>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Statistic title="Total Hits" value={linkStats.total_hits} />
            </Col>
            <Col xs={24} sm={8}>
              <Statistic title="Today's Hits" value={linkStats.today_hits} />
            </Col>
            <Col xs={24} sm={8}>
              <Statistic title="Unique Visitors" value={linkStats.unique_ips} />
            </Col>
          </Row>

          {linkStats.countries && linkStats.countries.length > 0 && (
            <div style={{ marginTop: '16px' }}>
              <h4>Country Breakdown:</h4>
              <Row gutter={[8, 8]}>
                {linkStats.countries.map((country, index) => (
                  <Col key={index}>
                    <Tag>{country.country}: {country.hits}</Tag>
                  </Col>
                ))}
              </Row>
            </div>
          )}

          {linkStats.targets && linkStats.targets.length > 0 && (
            <div style={{ marginTop: '16px' }}>
              <h4>Target Performance:</h4>
              <Row gutter={[8, 8]}>
                {linkStats.targets.map((target, index) => (
                  <Col key={index} xs={24}>
                    <Tag color="blue">
                      {target.url}: {target.hits} hits
                    </Tag>
                  </Col>
                ))}
              </Row>
            </div>
          )}
        </Card>
      )}
    </div>
  )
}

export default Statistics