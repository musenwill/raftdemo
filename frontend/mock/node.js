import Mock from 'mockjs'

const nodes = Mock.mock({
  'entries|3-5': [{
    'host|1': ['node-1', 'node-2', 'node-3', 'node-4'],
    'term|1': [1, 2, 3, 4],
    'state|1': ['Follower', 'Candidate', 'Leader'], 
    'commit_index|1': [-1, 0, 1, 2, 3],
    'last_applied_id|1': [-1, 0, 1, 2, 3],
    'leader|1': ['node-1', 'node2', 'node3', 'node4'],
    'vote_for|1': ['', '', '', '', '', '', '', 'node-1', 'node-2', 'node-3', 'node-4']
  }],
  'total|1-1': [2, 3, 4, 5],
})



export default [
  {
    url: '/v1/nodes',
    type: 'get',
    response: config => {
      return {
        total: nodes.total,
        entries: nodes.entries
      }
    }
  },
  {
    url: '/v1/nodes',
    type: 'put',
    response: config => {
      return {
      }
    }
  }
]
