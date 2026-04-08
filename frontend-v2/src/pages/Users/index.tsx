import React, { useRef, useState } from 'react';
import {
  ActionType,
  ModalForm,
  PageContainer,
  ProColumns,
  ProFormSelect,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import { Button, message, Popconfirm, Tag } from 'antd';
import { createUser, deleteUser, getUsers } from '@/services/user/api';

const fetchUsers = async (params: any) => {
  const { current = 1, pageSize = 10, ...rest } = params;

  // Build query parameters
  const queryParams = {
    page: current,
    limit: pageSize,
    ...rest,
  };

  try {
    const response = await getUsers(queryParams);

    if (response && response.data) {
      const tableData = response.data;

      tableData.users.map((item: User.User) => {
        return {
          key: item.id,
        };
      });

      return {
        data: tableData.users,
        success: true,
        total: response.data.totalComments,
      };
    }

    return {
      data: [],
      success: false,
    };
  } catch (error) {
    console.error('Failed to fetch users:', error);
    return {
      data: [],
      success: false,
    };
  }
};

const UserList: React.FC = () => {
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const actionRef = useRef<ActionType>();

  // Function to handle user creation
  const handleCreateUser = async (values: User.CreateUserRequest) => {
    try {
      const response = await createUser(values);

      if (response && response.code === 201) {
        message.success('User created successfully');
        // Refresh the table
        if (actionRef.current) {
          actionRef.current.reload();
        }
        return true; // Return true to close the modal
      } else {
        message.error('Failed to create user');
        return false; // Return false to keep the modal open
      }
    } catch (error) {
      console.error('Error creating user:', error);
      message.error('An error occurred while creating the user');
      return false; // Return false to keep the modal open
    }
  };

  // Function to handle user deletion
  const handleDeleteUser = async (userId: string) => {
    try {
      const response = await deleteUser({ id: userId });

      if (response && response.code === 200) {
        message.success('User deleted successfully');
        // Refresh the table
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error('Failed to delete user');
      }
    } catch (error) {
      console.error('Error deleting user:', error);
      message.error('An error occurred while deleting the user');
    }
  };

  const columns: ProColumns<User.User>[] = [
    {
      title: 'ID',
      dataIndex: '_id',
      hideInForm: true,
      hideInTable: true,
    },
    {
      title: 'Username',
      dataIndex: 'username',
      copyable: true,
      ellipsis: true,
    },
    {
      title: 'Role',
      dataIndex: 'role',
      render: (role) => <Tag color={role === 'Admin' ? 'red' : 'green'}>{role}</Tag>,
    },
    {
      title: 'Actions',
      key: 'action',
      width: 120,
      valueType: 'option',

      render: (_, record) => [
        <Popconfirm
          key="delete"
          title="Delete user"
          description="Are you sure you want to delete this user?"
          onConfirm={() => handleDeleteUser(record.id)}
          okText="Yes"
          cancelText="No"
        >
          <Button color="danger" key="delete" variant="text" size="small" danger>
            Delete
          </Button>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<User.User>
        request={fetchUsers}
        rowKey="_id"
        search={false}
        columns={columns}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setCreateModalVisible(true)}>
            Create User
          </Button>,
        ]}
      />
      {/* Create User Modal with ProForm */}
      <ModalForm<User.CreateUserRequest>
        title="Create User"
        width="400px"
        open={createModalVisible}
        onOpenChange={setCreateModalVisible}
        onFinish={handleCreateUser}
        submitter={{
          searchConfig: {
            submitText: 'Create',
            resetText: 'Cancel',
          },
        }}
      >
        <ProFormText
          name="username"
          label="Username"
          placeholder="Enter username"
          rules={[
            { required: true, message: 'Please enter username' },
            { min: 4, message: 'Username must be at least 4 characters' },
          ]}
        />

        <ProFormText.Password
          name="password"
          label="Password"
          placeholder="Enter password"
          rules={[
            { required: true, message: 'Please enter password' },
            { min: 6, message: 'Password must be at least 6 characters' },
          ]}
        />

        <ProFormSelect
          name="role"
          label="User Role"
          placeholder="Select role"
          options={[
            { label: 'Admin', value: 'Admin' },
            { label: 'User', value: 'User' },
          ]}
          rules={[{ required: true, message: 'Please select role' }]}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default UserList;
