import request from '@/utils/request'

export function GetNodes(params) {
  return request({
    url: '/v1/nodes',
    method: 'get',
    params
  })
}

export function UpdateNode(host, data) {
  return request({
    url: `/v1/nodes/${host}`,
    method: 'put',
    data
  })
}
