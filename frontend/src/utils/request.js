import axios from 'axios'
import { MessageBox, Message } from 'element-ui'

// create an axios instance
const service = axios.create({
  baseURL: process.env.VUE_APP_BASE_API, // url = base url + request url
  withCredentials: true, // send cookies when cross-domain requests
  timeout: 5000 // request timeout
})

// request interceptor
service.interceptors.request.use(
  config => {
    // do something before request is sent
    return config
  },
  error => {
    // do something with request error
    console.log(error) // for debug
    return Promise.reject(error)
  }
)

// response interceptor
service.interceptors.response.use(
  response => {
    // if the status code is not 200, it is judged as an error.
    if (response.status !== 200) {
      Message({
        status: response.status,
        code: response.data.code,
        message: response.data.error || 'error',
        type: 'error',
        duration: 5 * 1000
      })

      return Promise.reject(response || 'error')
    } else {
      return response.data
    }
  },
  error => {
    console.log(error)
    Message({
      status: error.response.status,
      code: error.response.data.code,
      message: error.response.data.error || 'error',
      type: 'error',
      duration: 5 * 1000
    })
    return Promise.reject(error)
  }
)

export default service
