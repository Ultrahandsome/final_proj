import { request } from '@umijs/max';

export async function getComments(params: Comment.GetCommentsRequest) {
  return request('/api/comments', {
    method: 'POST',
    data: { ...params },
  });
}

export async function getCategories() {
  return request('/api/categories');
}

// ✅ 新增：更新评论
export async function updateComment(data: { id: string; category: string; rawComment: string }) {
  const token = localStorage.getItem('token'); // 或从用户上下文获取

  return request('/api/comment/category', {
    method: 'POST',
    headers: {
      'X-Token': token || '',
    },
    data: {
      id: data.id,
      category: data.category,
      comment: data.rawComment, // ⚠️ 注意字段名对齐后端要求
    },
  });
}

// 下载 Excel 文件
export const exportCommentsExcel = async (ids: string[] = []) => {
  const response = await fetch('http://localhost:38080/api/export/excel', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Token': localStorage.getItem('token') || '',
    },
    body: JSON.stringify({ ids }),
  });
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', 'comments.xlsx');
  document.body.appendChild(link);
  link.click();
  link.remove();
};

// 下载 CSV 文件
export const exportCommentsCSV = async (ids: string[] = []) => {
  const response = await fetch('http://localhost:38080/api/export/csv', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Token': localStorage.getItem('token') || '',
    },
    body: JSON.stringify({ ids }),
  });
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', 'comments.csv');
  document.body.appendChild(link);
  link.click();
  link.remove();
};
