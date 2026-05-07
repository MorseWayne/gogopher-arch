import axios from 'axios';

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error)) {
      const message =
        typeof error.response?.data === 'string'
          ? error.response.data
          : error.message || '无法连接到 Gateway 服务';
      return Promise.reject(new Error(message));
    }
    return Promise.reject(error);
  }
);