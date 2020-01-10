import request from '@/utils/request'

export function GetLogs(params) {
  return request({
    url: '/v1/logs',
    method: 'get',
    params
  })
}

export function AddLog(data) {
  return request({
    url: `/v1/logs`,
    method: 'post',
    data
  })
}
