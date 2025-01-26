import { request } from '@umijs/max';
import {
  GetUserInfoApiRes,
  UpdateUserInfoApiParams,
  UpdateUserInfoApiRes,
  UserLoginApiParams,
  UserLoginApiRes,
} from '@/services/user/typings';
import { ApiResponse } from '@/services/typings';

const userLoginApi = async (data: UserLoginApiParams): Promise<ApiResponse<UserLoginApiRes>> => {
  const formData = new FormData();
  (Object.keys(data) as (keyof UserLoginApiParams)[]).forEach(key => formData.append(key as string, data[key]));

  return request('/api/v1/admin/login', {
    method: 'POST',
    data: formData,
    requestType: 'form',
  });
};

const updateUserInfoApi = async (data: UpdateUserInfoApiParams): Promise<ApiResponse<UpdateUserInfoApiRes>> => {
  return request('/api/v1/admin/updateUserInfo', {
    method: 'PATCH',
    data,
  });
};

const getUserInfoApi = async (): Promise<ApiResponse<GetUserInfoApiRes>> => {
  return request('/api/v1/admin/userInfo', {
    method: 'GET',
  });
};

export default {
  userLoginApi,
  updateUserInfoApi,
  getUserInfoApi,
};
