import React, { useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
  Tag,
  Tooltip,
  Row,
  Col,
  Avatar,
  Descriptions,
  Transfer,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  LinkOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/services/api'
import type { ColumnsType } from 'antd/es/table'
import type { TransferProps } from 'antd'

const { Option } = Select

interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'editor' | 'viewer'
  is_active: boolean
  last_login: string
  created_at: string
  updated_at: string
}

interface CreateUserRequest {
  username: string
  email: string
  password: string
  role: string
  is_active: boolean
}

interface Link {
  id: number
  link_id: string
  business_unit: string
  network: string
}

interface TransferItem {
  key: string
  title: string
  description: string
}

const Users: React.FC = () => {
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false)
  const [isLinksModalVisible, setIsLinksModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [form] = Form.useForm()
  const [linksForm] = Form.useForm()
  const queryClient = useQueryClient()

  // Fetch users
  const { data: usersData, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => api.get('/users').then(res => res.data),
  })

  // Fetch all links for assignment
  const { data: allLinksData } = useQuery({
    queryKey: ['allLinks'],
    queryFn: () => api.get('/links').then(res => res.data),
  })

  // Fetch user's assigned links
  const { data: userLinksData } = useQuery({
    queryKey: ['userLinks', selectedUser?.id],
    queryFn: () => api.get(`/users/${selectedUser?.id}/links`).then(res => res.data),
    enabled: !!selectedUser,
  })

  // Create user mutation
  const createUserMutation = useMutation({
    mutationFn: (data: CreateUserRequest) => api.post('/users', data),
    onSuccess: () => {
      message.success('User created successfully')
      setIsCreateModalVisible(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: () => {
      message.error('Failed to create user')
    },
  })

  // Update user mutation
  const updateUserMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<CreateUserRequest> }) =>
      api.put(`/users/${id}`, data),
    onSuccess: () => {
      message.success('User updated successfully')
      setEditingUser(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: () => {
      message.error('Failed to update user')
    },
  })

  // Delete user mutation
  const deleteUserMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/users/${id}`),
    onSuccess: () => {
      message.success('User deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: () => {
      message.error('Failed to delete user')
    },
  })

  // Assign links mutation
  const assignLinksMutation = useMutation({
    mutationFn: ({ userId, linkIds }: { userId: number; linkIds: number[] }) =>
      Promise.all(
        linkIds.map(linkId =>
          api.post(`/users/${userId}/links`, { link_id: linkId })
        )
      ),
    onSuccess: () => {
      message.success('Links assigned successfully')
      queryClient.invalidateQueries({ queryKey: ['userLinks'] })
    },
    onError: () => {
      message.error('Failed to assign links')
    },
  })

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      
      if (editingUser) {
        // Don't send password if it's empty on edit
        const updateData = { ...values }
        if (!updateData.password) {
          delete updateData.password
        }
        updateUserMutation.mutate({ id: editingUser.id, data: updateData })
      } else {
        createUserMutation.mutate(values)
      }
    } catch (error) {
      console.error('Form validation failed:', error)
    }
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    form.setFieldsValue({
      ...user,
      password: '', // Don't populate password field
    })
    setIsCreateModalVisible(true)
  }

  const handleManageLinks = (user: User) => {
    setSelectedUser(user)
    setIsLinksModalVisible(true)
  }

  const handleLinksAssignment = (targetKeys: string[]) => {
    if (selectedUser) {
      const linkIds = targetKeys.map(key => parseInt(key))
      assignLinksMutation.mutate({ userId: selectedUser.id, linkIds })
    }
  }

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'red'
      case 'editor':
        return 'blue'
      case 'viewer':
        return 'green'
      default:
        return 'default'
    }
  }

  const columns: ColumnsType<User> = [
    {
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      render: (text: string, record: User) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <Avatar icon={<UserOutlined />} />
          <div>
            <div style={{ fontWeight: 600 }}>{text}</div>
            <div style={{ fontSize: '12px', color: '#666' }}>{record.email}</div>
          </div>
        </div>
      ),
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={getRoleColor(role)}>
          {role.toUpperCase()}
        </Tag>
      ),
      filters: [
        { text: 'Admin', value: 'admin' },
        { text: 'Editor', value: 'editor' },
        { text: 'Viewer', value: 'viewer' },
      ],
      onFilter: (value, record) => record.role === value,
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
      filters: [
        { text: 'Active', value: true },
        { text: 'Inactive', value: false },
      ],
      onFilter: (value, record) => record.is_active === value,
    },
    {
      title: 'Last Login',
      dataIndex: 'last_login',
      key: 'last_login',
      render: (date: string) =>
        date ? new Date(date).toLocaleString() : 'Never',
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record: User) => (
        <Space>
          <Tooltip title="Manage Links">
            <Button
              type="primary"
              icon={<LinkOutlined />}
              size="small"
              onClick={() => handleManageLinks(record)}
            />
          </Tooltip>
          <Tooltip title="Edit User">
            <Button
              icon={<EditOutlined />}
              size="small"
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="Are you sure you want to delete this user?"
            onConfirm={() => deleteUserMutation.mutate(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              danger
              icon={<DeleteOutlined />}
              size="small"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // Prepare transfer data
  const transferData: TransferItem[] = allLinksData?.data?.map((link: Link) => ({
    key: link.id.toString(),
    title: `${link.link_id} - ${link.business_unit}`,
    description: link.network,
  })) || []

  const assignedLinkIds = userLinksData?.data?.map((link: Link) => link.id.toString()) || []

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <h1>Users Management</h1>
            <p>Manage user accounts and permissions</p>
          </Col>
          <Col>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditingUser(null)
                form.resetFields()
                setIsCreateModalVisible(true)
              }}
            >
              Create User
            </Button>
          </Col>
        </Row>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={usersData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} users`,
          }}
        />
      </Card>

      {/* Create/Edit User Modal */}
      <Modal
        title={editingUser ? 'Edit User' : 'Create User'}
        open={isCreateModalVisible}
        onCancel={() => {
          setIsCreateModalVisible(false)
          setEditingUser(null)
          form.resetFields()
        }}
        onOk={handleSubmit}
        confirmLoading={createUserMutation.isPending || updateUserMutation.isPending}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            role: 'viewer',
            is_active: true,
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="username"
                label="Username"
                rules={[{ required: true, message: 'Please enter username' }]}
              >
                <Input placeholder="Enter username" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="email"
                label="Email"
                rules={[
                  { required: true, message: 'Please enter email' },
                  { type: 'email', message: 'Please enter valid email' },
                ]}
              >
                <Input placeholder="user@example.com" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="password"
            label="Password"
            rules={[
              {
                required: !editingUser,
                message: 'Please enter password',
              },
              {
                min: 6,
                message: 'Password must be at least 6 characters',
              },
            ]}
          >
            <Input.Password
              placeholder={editingUser ? 'Leave empty to keep current password' : 'Enter password'}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="role"
                label="Role"
                rules={[{ required: true, message: 'Please select role' }]}
              >
                <Select placeholder="Select role">
                  <Option value="viewer">Viewer</Option>
                  <Option value="editor">Editor</Option>
                  <Option value="admin">Admin</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="is_active"
                label="Status"
                valuePropName="checked"
              >
                <Select placeholder="Select status">
                  <Option value={true}>Active</Option>
                  <Option value={false}>Inactive</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      {/* Manage Links Modal */}
      <Modal
        title={`Manage Links - ${selectedUser?.username}`}
        open={isLinksModalVisible}
        onCancel={() => {
          setIsLinksModalVisible(false)
          setSelectedUser(null)
        }}
        footer={null}
        width={800}
      >
        {selectedUser && (
          <div>
            <Descriptions size="small" style={{ marginBottom: '16px' }}>
              <Descriptions.Item label="User">{selectedUser.username}</Descriptions.Item>
              <Descriptions.Item label="Role">{selectedUser.role}</Descriptions.Item>
              <Descriptions.Item label="Email">{selectedUser.email}</Descriptions.Item>
            </Descriptions>

            <Transfer
              dataSource={transferData}
              titles={['Available Links', 'Assigned Links']}
              targetKeys={assignedLinkIds}
              onChange={handleLinksAssignment}
              render={item => `${item.title} - ${item.description}`}
              listStyle={{
                width: 350,
                height: 400,
              }}
              operations={['Assign', 'Unassign']}
              showSearch
              filterOption={(inputValue, option) =>
                option.title.toLowerCase().includes(inputValue.toLowerCase()) ||
                option.description.toLowerCase().includes(inputValue.toLowerCase())
              }
            />
          </div>
        )}
      </Modal>
    </div>
  )
}

export default Users