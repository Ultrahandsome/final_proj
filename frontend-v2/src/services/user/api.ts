import { request } from '@umijs/max';

export async function getUsers(params: User.GetUsersRequest) {
  return request('/api/user', {
    method: 'POST',
    data: { ...params },
  });
}

export async function createUser(params: User.CreateUserRequest) {
  return request('/api/user/add', {
    method: 'POST',
    data: { ...params },
  });
}

export async function deleteUser(params: User.DeleteUserRequest) {
  return request('/api/user/delete', {
    method: 'POST',
    data: { ...params },
  });
}

export async function getUserInfo() {
  const token = localStorage.getItem('token') || '';

  return request('/api/info', {
    method: 'GET',
    headers: {
      'X-Token': token,
    },
  });
}
