import React, { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Descriptions, Tag, Button, Space, Table, Modal, message, Tabs, Statistic, Row, Col } from 'antd'
import {
  EditOutlined,
  PlusOutlined,
  DeleteOutlined,
  CopyOutlined,
  BarChartOutlined,
  LinkOutlined,
  GlobalOutlined,
} from '@ant-design/icons'
import { useLink, useTargets, useLinkStats, useDeleteTarget } from '@/hooks/useApi'
import { formatNumber, formatDate, generateShortUrl, copyToClipboard, getCountryFlag } from '@/utils/format'
import CreateTargetModal from './CreateTargetModal'
import type { Target } from '@/types/api'

const LinkDetail: React.FC = () => {
  const { linkId } = useParams<{ linkId: string }>()
  const navigate = useNavigate()
  const [createTargetVisible, setCreateTargetVisible] = useState(false)
  
  const { data: link } = useLink(linkId!)
  const { data: targets } = useTargets(linkId!)
  const { data: stats } = useLinkStats(linkId!)
  const deleteTargetMutation = useDeleteTarget()

  const handleCopyUrl = async () => {
    if (!link) return
    const url = generateShortUrl(link.business_unit, link.link_id)
    const success = await copyToClipboard(url)
    if (success) {
      message.success('URL copied to clipboard!')
    }
  }

  const handleDeleteTarget = (targetId: number) => {
    Modal.confirm({
      title: 'Delete Target',
      content: 'Are you sure you want to delete this target?',
      okText: 'Delete',
      okType: 'danger',
      onOk: () => {
        deleteTargetMutation.mutate(targetId)
      },
    })
  }

  const targetColumns = [
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      render: (url: string) => (
        <a href={url} target="_blank" rel="noopener noreferrer">
          {url}
        </a>
      ),
    },
    {
      title: 'Weight',
      dataIndex: 'weight',
      key: 'weight',
      width: 100,
      render: (weight: number) => <Tag color="blue">{weight}%</Tag>,
    },
    {
      title: 'Hits / Cap',
      key: 'hits',
      width: 120,
      render: (_: any, record: Target) => (
        <Space size={4}>
          <span>{formatNumber(record.current_hits)}</span>
          <span>/</span>
          <span>{record.cap > 0 ? formatNumber(record.cap) : '∞'}</span>
        </Space>
      ),
    },
    {
      title: 'Countries',
      dataIndex: 'countries',
      key: 'countries',
      width: 200,
      render: (countries: string | string[]) => {
        // Handle both string and array formats
        let countryList: string[] = []
        if (typeof countries === 'string') {
          try {
            countryList = JSON.parse(countries)
          } catch {
            countryList = []
          }
        } else if (Array.isArray(countries)) {
          countryList = countries
        }
        
        if (!countryList.length) {
          return <Tag color="default">All Countries</Tag>
        }
        
        return (
          <Space size={4} wrap>
            {countryList.map(country => (
              <Tag key={country} icon={getCountryFlag(country)}>
                {country}
              </Tag>
            ))}
          </Space>
        )
      },
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'success' : 'default'}>
          {isActive ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: any, record: Target) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => navigate(`/targets/${record.id}/edit`)}
          >
            Edit
          </Button>
          <Button
            type="link"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteTarget(record.id)}
          >
            Delete
          </Button>
        </Space>
      ),
    },
  ]

  const items = [
    {
      key: 'details',
      label: 'Details',
      icon: <LinkOutlined />,
      children: (
        <Card>
          <Descriptions column={2} bordered>
            <Descriptions.Item label="Link ID">
              <code style={{ background: '#f6f8fa', padding: '2px 6px', borderRadius: '4px' }}>
                {link?.link_id}
              </code>
            </Descriptions.Item>
            <Descriptions.Item label="Business Unit">
              <Tag color="blue">{link?.business_unit}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Network">
              <Tag color="green">{link?.network}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Status">
              <Tag color={link?.is_active ? 'success' : 'default'}>
                {link?.is_active ? 'Active' : 'Inactive'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Total Cap">
              {link?.total_cap ? formatNumber(link.total_cap) : 'Unlimited'}
            </Descriptions.Item>
            <Descriptions.Item label="Current Hits">
              {formatNumber(link?.current_hits || 0)}
            </Descriptions.Item>
            <Descriptions.Item label="Backup URL" span={2}>
              {link?.backup_url || 'Not configured'}
            </Descriptions.Item>
            <Descriptions.Item label="Short URL" span={2}>
              <Space>
                <code>
                  https://{link && generateShortUrl(link.business_unit, link.link_id, link.network)}
                </code>
                <Button
                  type="link"
                  icon={<CopyOutlined />}
                  onClick={handleCopyUrl}
                >
                  Copy
                </Button>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="Created">
              {link && formatDate(link.created_at)}
            </Descriptions.Item>
            <Descriptions.Item label="Updated">
              {link && formatDate(link.updated_at)}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'targets',
      label: 'Targets',
      icon: <GlobalOutlined />,
      children: (
        <Card
          extra={
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateTargetVisible(true)}
            >
              Add Target
            </Button>
          }
        >
          <Table
            columns={targetColumns}
            dataSource={targets}
            rowKey="id"
            pagination={false}
          />
        </Card>
      ),
    },
    {
      key: 'statistics',
      label: 'Statistics',
      icon: <BarChartOutlined />,
      children: stats ? (
        <div>
          <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="Total Hits"
                  value={stats.total_hits}
                  prefix={<LinkOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="Today's Hits"
                  value={stats.today_hits}
                  valueStyle={{ color: '#3f8600' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="Unique IPs"
                  value={stats.unique_ips}
                  prefix={<GlobalOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="Countries"
                  value={stats.countries?.length || 0}
                />
              </Card>
            </Col>
          </Row>

          <Card title="Traffic by Country">
            <Table
              dataSource={stats.countries}
              columns={[
                {
                  title: 'Country',
                  dataIndex: 'country',
                  key: 'country',
                  render: (country: string) => (
                    <Space>
                      {getCountryFlag(country)}
                      {country}
                    </Space>
                  ),
                },
                {
                  title: 'Hits',
                  dataIndex: 'hits',
                  key: 'hits',
                  render: (hits: number) => formatNumber(hits),
                },
                {
                  title: 'Percentage',
                  key: 'percentage',
                  render: (_, record) => (
                    <span>
                      {((record.hits / stats.total_hits) * 100).toFixed(1)}%
                    </span>
                  ),
                },
              ]}
              pagination={false}
            />
          </Card>
        </div>
      ) : null,
    },
  ]

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ margin: 0 }}>Link Details</h2>
          <Space>
            <Button onClick={() => navigate('/links')}>Back to Links</Button>
            <Button
              type="primary"
              icon={<EditOutlined />}
              onClick={() => navigate(`/links/${linkId}/edit`)}
            >
              Edit Link
            </Button>
          </Space>
        </div>
      </Card>

      <Tabs items={items} />

      <CreateTargetModal
        visible={createTargetVisible}
        linkId={linkId!}
        onClose={() => setCreateTargetVisible(false)}
      />
    </div>
  )
}

export default LinkDetail