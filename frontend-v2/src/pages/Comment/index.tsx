import { useRef, useState } from 'react';
import {
  PageContainer,
  ProTable,
  ModalForm,
  ProFormText,
} from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Tag, Button, Divider, message, Form } from 'antd';
import dayjs from 'dayjs';
import { getComments, updateComment } from '@/services/comments/api';
import ExportFile from '@/components/ExportFile';

interface CommentTableItem {
  key: string;
  id: string;
  rawComment: string;
  category: string;
  confidenceScore: string;
  keywords: string[];
  lastUpdated: string;
  lastUpdatedBy: string;
  deleted: boolean;
  updateHistory: Comment.UpdateHistory[];
  similarComments: string[];
}

const transformToTableData = (apiData: any): CommentTableItem[] => {
  if (!apiData || !apiData.comments || !Array.isArray(apiData.comments)) {
    return [];
  }

  return apiData.comments.map((item: Comment.Comment) => ({
    key: item._id,
    id: item._id,
    rawComment: item.rawComment,
    category: item.category,
    confidenceScore: (item.confidenceScore * 100).toFixed(2) + '%',
    keywords: item.keywords || [],
    lastUpdated: item.lastUpdated
      ? dayjs(item.lastUpdated * 1000).format('YYYY-MM-DD HH:mm:ss')
      : '',
    lastUpdatedBy: item.lastUpdated || '',
    deleted: item.isDeleted || false,
    updateHistory: item.updateHistory || [],
    similarComments: item.similarComments || [],
  }));
};

export default () => {
  const actionRef = useRef<ActionType>();
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editingComment, setEditingComment] = useState<CommentTableItem | null>(null);
  const [form] = Form.useForm();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [categoryOptions, setCategoryOptions] = useState<{ text: string; value: string }[]>([]);

  // Function to extract unique categories from data
  const extractCategories = (data: CommentTableItem[]) => {
    const categories = new Set<string>();
    data.forEach((item) => {
      if (item.category) {
        categories.add(item.category);
      }
    });
    return Array.from(categories).map(category => ({
      text: category,
      value: category,
    }));
  };

  const handleEdit = (record: CommentTableItem) => {
    setEditingComment(record);
    form.setFieldsValue({
      category: record.category,
      rawComment: '',
    });
    setEditModalVisible(true);
  };

  const handleEditSubmit = async (values: any) => {
    if (!editingComment) return false;

    try {
      const response = await updateComment({
        id: editingComment.id,
        category: values.category,
        rawComment: values.rawComment,
      });

      console.log('update response', response);

      if (response?.code === 200 || response?.code === 201 || response?.success === true) {
        message.success('Comment updated successfully');
        setEditModalVisible(false);
        actionRef.current?.reloadAndRest?.();
        return true;
      } else {
        message.error('Update failed');
        return false;
      }
    } catch (error) {
      console.error('Error updating comment:', error);
      message.error('An error occurred while updating the comment');
      return false;
    }
  };

  const columns: ProColumns<CommentTableItem>[] = [
    {
      title: 'ID',
      dataIndex: '_id',
      hideInForm: true,
      hideInTable: true,
    },
    {
      title: 'Comment',
      dataIndex: 'rawComment',
      copyable: true,
      ellipsis: true,
      search: true,
      width: 300,
    },
    {
      title: 'Category',
      dataIndex: 'category',
      width: 120,
      filters: categoryOptions,
      filterMultiple: true,
      onFilter: (value, record) => record.category === value,
      valueEnum: categoryOptions.reduce((acc, { value, text }) => {
        acc[value] = { text };
        return acc;
      }, {}),
    },
    {
      title: 'Confidence Score',
      dataIndex: 'confidenceScore',
      width: 100,
      sorter: (a, b) => parseFloat(a.confidenceScore) - parseFloat(b.confidenceScore),
    },
    {
      title: 'Last Updated',
      dataIndex: 'lastUpdated',
      width: 170,
      sorter: (a, b) => dayjs(a.lastUpdated).unix() - dayjs(b.lastUpdated).unix(),
    },
    {
      title: 'Actions',
      key: 'action',
      width: 120,
      valueType: 'option',
      render: (_, record) => [
        <Button key="edit" type="link" size="small" onClick={() => handleEdit(record)}>
          Edit Category
        </Button>,
      ],
    },
  ];

  const fetchComments = async (params: any) => {
    const { current = 1, pageSize = 10, ...rest } = params;

    try {
      const response = await getComments({
        page: current,
        limit: pageSize,
        ...rest,
      });

      if (response && response.data) {
        const tableData = transformToTableData(response.data);

        // Update category options when data is fetched
        const newCategoryOptions = extractCategories(tableData);
        setCategoryOptions(newCategoryOptions);

        return {
          data: tableData,
          success: true,
          total: response.data.totalComments,
        };
      }

      return { data: [], success: false };
    } catch (error) {
      console.error('Failed to fetch comments:', error);
      return { data: [], success: false };
    }
  };

  return (
    <PageContainer>
      <ProTable<CommentTableItem>
        headerTitle="Comments List"
        actionRef={actionRef}
        rowKey="key"
        search={false}
        request={fetchComments}
        columns={columns}
        rowSelection={{ selectedRowKeys, onChange: setSelectedRowKeys }}
        pagination={{
          showQuickJumper: true,
          defaultPageSize: 10,
          showSizeChanger: true,
        }}
        toolBarRender={() => [
          <ExportFile key="export" selectedIds={selectedRowKeys as string[]} />,
        ]}
        expandable={{
          expandedRowRender: (record) => (
            <>
              <div style={{ margin: 0 }}>
                <h4>Content</h4>
                <div>{record.rawComment}</div>
              </div>
              <Divider />
              <div style={{ margin: 0 }}>
                <h4>Keywords</h4>
                <ul>
                  {record.keywords?.map((keyword, index) => (
                    <Tag key={index} color="blue">
                      {keyword}
                    </Tag>
                  ))}
                </ul>
              </div>
              <div style={{ margin: 0 }}>
                <h4>Update History</h4>
                <ul>
                  {record.updateHistory?.map((update, index) => (
                    <li key={index}>
                      {dayjs(update.time * 1000).format('YYYY-MM-DD HH:mm:ss')} - User{' '}
                      <b>{update.user}</b> changed to <b>{update.category}</b>, Note:{' '}
                      <b>{update.comment || 'N/A'}</b>
                    </li>
                  ))}
                </ul>
              </div>
              <div style={{ margin: 0 }}>
                <h4>Similar Comments</h4>
                <ul>
                  {record.similarComments?.map((similarComment, index) => (
                    <li key={index}>{similarComment}</li>
                  ))}
                </ul>
              </div>
            </>
          ),
        }}
      />

      <ModalForm
        form={form}
        title="Edit Comment"
        width="400px"
        open={editModalVisible}
        onOpenChange={setEditModalVisible}
        onFinish={handleEditSubmit}
      >
        <ProFormText
          name="rawComment"
          label="Note"
          placeholder="Edit comment"
          rules={[{ required: true, message: 'Please enter note text' }]}
        />
        <ProFormText
          name="category"
          label="Category"
          placeholder="Enter category (e.g. Positive, Negative, Neutral)"
          rules={[{ required: true, message: 'Please enter a category' }]}
        />
      </ModalForm>
    </PageContainer>
  );
};
