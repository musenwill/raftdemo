import request from '@/utils/request'

export function GetConfig(params) {
  return request({
    url: 'v1/config',
    method: 'get',
    params
  })
}

export function UpdateConfig(data) {
  return request({
    url: 'v1/config',
    method: 'put',
    data
  })
}
