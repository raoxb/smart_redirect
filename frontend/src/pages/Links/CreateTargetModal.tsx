import React from 'react'
import { Modal, Form, Input, InputNumber, Select, Button, Space } from 'antd'
import { useCreateTarget } from '@/hooks/useApi'
import type { CreateTargetRequest } from '@/types/api'

interface CreateTargetModalProps {
  visible: boolean
  linkId: string
  onClose: () => void
}

const countryOptions = [
  { label: '🇺🇸 United States', value: 'US' },
  { label: '🇨🇦 Canada', value: 'CA' },
  { label: '🇬🇧 United Kingdom', value: 'UK' },
  { label: '🇩🇪 Germany', value: 'DE' },
  { label: '🇫🇷 France', value: 'FR' },
  { label: '🇯🇵 Japan', value: 'JP' },
  { label: '🇨🇳 China', value: 'CN' },
  { label: '🇮🇳 India', value: 'IN' },
  { label: '🇧🇷 Brazil', value: 'BR' },
  { label: '🇦🇺 Australia', value: 'AU' },
]

const CreateTargetModal: React.FC<CreateTargetModalProps> = ({ visible, linkId, onClose }) => {
  const [form] = Form.useForm()
  const createMutation = useCreateTarget()

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data: CreateTargetRequest = {
        url: values.url,
        weight: values.weight,
        cap: values.cap || 0,
        countries: values.countries || [],
        param_mapping: values.param_mapping || {},
        static_params: values.static_params || {},
      }
      
      await createMutation.mutateAsync({ linkId, data })
      form.resetFields()
      onClose()
    } catch (error) {
      // Error handling is done in the hook
    }
  }

  return (
    <Modal
      title="Add Target"
      open={visible}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={createMutation.isPending}
      width={700}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          weight: 100,
          cap: 0,
        }}
      >
        <Form.Item
          name="url"
          label="Target URL"
          rules={[
            { required: true, message: 'Please enter target URL' },
            { type: 'url', message: 'Please enter a valid URL' },
          ]}
        >
          <Input placeholder="https://example.com/landing" />
        </Form.Item>

        <Form.Item
          name="weight"
          label="Weight"
          tooltip="Traffic distribution weight (relative to other targets)"
          rules={[{ required: true, message: 'Please enter weight' }]}
        >
          <InputNumber
            min={1}
            max={1000}
            style={{ width: '100%' }}
            addonAfter="%"
          />
        </Form.Item>

        <Form.Item
          name="cap"
          label="Cap (0 for unlimited)"
          tooltip="Maximum number of hits for this target"
        >
          <InputNumber
            min={0}
            style={{ width: '100%' }}
            placeholder="Enter 0 for unlimited"
          />
        </Form.Item>

        <Form.Item
          name="countries"
          label="Allowed Countries"
          tooltip="Leave empty to allow all countries"
        >
          <Select
            mode="multiple"
            placeholder="Select allowed countries"
            options={countryOptions}
            allowClear
          />
        </Form.Item>

        <Form.Item label="Parameter Mapping" tooltip="Map original parameters to new names">
          <Form.List name="param_mapping">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'from']}
                      rules={[{ required: true, message: 'Missing parameter' }]}
                    >
                      <Input placeholder="Original param" />
                    </Form.Item>
                    <span>→</span>
                    <Form.Item
                      {...restField}
                      name={[name, 'to']}
                      rules={[{ required: true, message: 'Missing parameter' }]}
                    >
                      <Input placeholder="New param" />
                    </Form.Item>
                    <Button type="link" onClick={() => remove(name)}>
                      Remove
                    </Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block>
                    Add Parameter Mapping
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form.Item>

        <Form.Item label="Static Parameters" tooltip="Add static parameters to all redirects">
          <Form.List name="static_params">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: 'Missing key' }]}
                    >
                      <Input placeholder="Parameter name" />
                    </Form.Item>
                    <span>=</span>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: 'Missing value' }]}
                    >
                      <Input placeholder="Parameter value" />
                    </Form.Item>
                    <Button type="link" onClick={() => remove(name)}>
                      Remove
                    </Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block>
                    Add Static Parameter
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default CreateTargetModal