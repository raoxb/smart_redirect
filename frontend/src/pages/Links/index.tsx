import React, { useState } from 'react'
import { Card, Table, Button, Space, Tag, Input, Select, Modal, message, Dropdown } from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  EyeOutlined,
  MoreOutlined,
  SearchOutlined,
  DownloadOutlined,
  UploadOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useLinks, useDeleteLink } from '@/hooks/useApi'
import { formatNumber, formatDate, generateShortUrl, copyToClipboard } from '@/utils/format'
import CreateLinkModal from './CreateLinkModal'
import QuickCreateModal from './QuickCreateModal'
import ImportModal from './ImportModal'
import type { Link } from '@/types/api'

const { Search } = Input

const LinksPage: React.FC = () => {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [searchText, setSearchText] = useState('')
  const [selectedBU, setSelectedBU] = useState<string>()
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [quickCreateModalVisible, setQuickCreateModalVisible] = useState(false)
  const [importModalVisible, setImportModalVisible] = useState(false)
  
  const { data, isLoading } = useLinks(page, pageSize)
  const deleteMutation = useDeleteLink()

  const handleDelete = (linkId: string) => {
    Modal.confirm({
      title: 'Delete Link',
      content: 'Are you sure you want to delete this link? This action cannot be undone.',
      okText: 'Delete',
      okType: 'danger',
      onOk: () => {
        deleteMutation.mutate(linkId)
      },
    })
  }

  const handleCopyUrl = async (link: Link) => {
    const url = generateShortUrl(link.business_unit, link.link_id, link.network)
    const success = await copyToClipboard(`https://${url}`)
    if (success) {
      message.success('URL copied to clipboard!')
    } else {
      message.error('Failed to copy URL')
    }
  }

  const handleExport = () => {
    // Implementation for export
    message.info('Export feature coming soon')
  }

  const columns = [
    {
      title: 'Link ID',
      dataIndex: 'link_id',
      key: 'link_id',
      fixed: 'left' as const,
      width: 120,
      render: (linkId: string) => (
        <code style={{ 
          background: '#f6f8fa', 
          padding: '2px 6px', 
          borderRadius: '4px',
          fontSize: '12px'
        }}>
          {linkId}
        </code>
      ),
    },
    {
      title: 'Business Unit',
      dataIndex: 'business_unit',
      key: 'business_unit',
      width: 120,
      filters: [
        { text: 'bu01', value: 'bu01' },
        { text: 'bu02', value: 'bu02' },
      ],
      render: (bu: string) => <Tag color="blue">{bu}</Tag>,
    },
    {
      title: 'Network',
      dataIndex: 'network',
      key: 'network',
      width: 100,
      render: (network: string) => <Tag color="green">{network}</Tag>,
    },
    {
      title: 'Hits / Cap',
      key: 'hits',
      width: 120,
      render: (_: any, record: Link) => (
        <Space size={4}>
          <span>{formatNumber(record.current_hits)}</span>
          <span>/</span>
          <span>{record.total_cap > 0 ? formatNumber(record.total_cap) : '∞'}</span>
        </Space>
      ),
    },
    {
      title: 'Progress',
      key: 'progress',
      width: 150,
      render: (_: any, record: Link) => {
        const percentage = record.total_cap > 0 
          ? Math.round((record.current_hits / record.total_cap) * 100)
          : 0
        return (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <div style={{ 
              flex: 1, 
              height: 8, 
              background: '#f0f0f0', 
              borderRadius: 4,
              overflow: 'hidden'
            }}>
              <div style={{
                width: `${percentage}%`,
                height: '100%',
                background: percentage > 80 ? '#ff4d4f' : '#52c41a',
                transition: 'width 0.3s ease'
              }} />
            </div>
            <span style={{ fontSize: 12, color: '#666' }}>{percentage}%</span>
          </div>
        )
      },
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 100,
      filters: [
        { text: 'Active', value: true },
        { text: 'Inactive', value: false },
      ],
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'success' : 'default'}>
          {isActive ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => formatDate(date),
      sorter: true,
    },
    {
      title: 'Actions',
      key: 'actions',
      fixed: 'right' as const,
      width: 120,
      render: (_: any, record: Link) => {
        const menuItems = [
          {
            key: 'view',
            icon: <EyeOutlined />,
            label: 'View Details',
            onClick: () => navigate(`/links/${record.link_id}`),
          },
          {
            key: 'edit',
            icon: <EditOutlined />,
            label: 'Edit',
            onClick: () => navigate(`/links/${record.link_id}/edit`),
          },
          {
            key: 'copy',
            icon: <CopyOutlined />,
            label: 'Copy URL',
            onClick: () => handleCopyUrl(record),
          },
          {
            type: 'divider' as const,
          },
          {
            key: 'delete',
            icon: <DeleteOutlined />,
            label: 'Delete',
            danger: true,
            onClick: () => handleDelete(record.link_id),
          },
        ]

        return (
          <Dropdown menu={{ items: menuItems }} trigger={['click']}>
            <Button type="text" icon={<MoreOutlined />} />
          </Dropdown>
        )
      },
    },
  ]

  const filteredData = data?.data?.filter(link => {
    if (searchText && !link.link_id.toLowerCase().includes(searchText.toLowerCase())) {
      return false
    }
    if (selectedBU && link.business_unit !== selectedBU) {
      return false
    }
    return true
  })

  return (
    <div>
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: 16 }}>
          <Space wrap>
            <Search
              placeholder="Search by link ID"
              allowClear
              enterButton={<SearchOutlined />}
              style={{ width: 250 }}
              onSearch={setSearchText}
            />
            <Select
              placeholder="Business Unit"
              allowClear
              style={{ width: 150 }}
              onChange={setSelectedBU}
              options={[
                { label: 'All', value: undefined },
                { label: 'bu01', value: 'bu01' },
                { label: 'bu02', value: 'bu02' },
              ]}
            />
          </Space>
          
          <Space>
            <Button 
              icon={<UploadOutlined />}
              onClick={() => setImportModalVisible(true)}
            >
              Import
            </Button>
            <Button 
              icon={<DownloadOutlined />}
              onClick={handleExport}
            >
              Export
            </Button>
            <Dropdown.Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setQuickCreateModalVisible(true)}
              menu={{
                items: [
                  {
                    key: 'quick',
                    label: 'Quick Create (with targets)',
                    onClick: () => setQuickCreateModalVisible(true),
                  },
                  {
                    key: 'basic',
                    label: 'Create Link Only',
                    onClick: () => setCreateModalVisible(true),
                  },
                ],
              }}
            >
              Quick Create
            </Dropdown.Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={filteredData}
          loading={isLoading}
          rowKey="id"
          scroll={{ x: 1200 }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: data?.total || 0,
            showSizeChanger: true,
            showTotal: (total) => `Total ${total} links`,
            onChange: (p, ps) => {
              setPage(p)
              setPageSize(ps || 20)
            },
          }}
        />
      </Card>

      <CreateLinkModal
        visible={createModalVisible}
        onClose={() => setCreateModalVisible(false)}
      />

      <QuickCreateModal
        visible={quickCreateModalVisible}
        onClose={() => setQuickCreateModalVisible(false)}
      />

      <ImportModal
        visible={importModalVisible}
        onClose={() => setImportModalVisible(false)}
      />
    </div>
  )
}

export default LinksPage