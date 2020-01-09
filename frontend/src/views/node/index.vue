<template>
  <div class="entries-text-test">
    <div class="entries-table">
      <el-table :data="nodes.entries" border fit highlight-current-row :header-cell-style="tableHeaderColor">
        <el-table-column align="center" prop="host" label="host" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="term" label="term" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="state" label="state" min-width="15%">
          <template slot-scope="{row}">
            <el-tag :type="row.state | stateFilter" min-width="10%">
              {{ row.state }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="center" prop="leader" label="leader" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="commit_index" label="commit index" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="last_applied_id" label="last applied id" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="vote_for" label="vote for" min-width="15%">
        </el-table-column>
        <el-table-column label="Actions" align="center" width="200" class-name="small-padding fixed-width">
          <template slot-scope="{row}">
            <el-button type="primary" size="mini" @click="sleepHandler(row)">
              sleep
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>
<script>

import { GetNodes, UpdateNode } from '@/api/node'

export default {
  name: 'Node',
  timer: '',
  filters: {
    stateFilter(state) {
      const stateMap = {
        Leader: 'success',
        Candidate: 'warning',
        Follower: 'primary'
      }
      return stateMap[state]
    }
  },
  data() {
    return {
      nodes: {
        total: 0,
        entries: []
      }
    }
  },
  created() {
    this.fetchData()
    this.timer = setInterval(this.fetchData, 500)
  },
  beforeDestroy () {
    clearInterval(this.timer)
  },
  methods: {
    fetchData() {
      GetNodes().then(response => {
        if (response != null) {
          this.nodes = response
        }
      })
    },
    tableHeaderColor({ row, column, rowIndex, columnIndex }) {
      if (rowIndex === 0) {
        return 'background-color: gray; color: #fff; font-weight: 500;'
      }
    },
    sleepHandler(row) {
      const param = {
        sleep: true
      }
      UpdateNode(row.host, param).then(() => {
        this.$notify({
          title: 'Success',
          message: 'Successfully',
          type: 'success',
          duration: 2000
        })
        fetchData()
      })
    },
    cancelAutoUpdate () {
      clearInterval(this.timer)
    }
  }
}
</script>

<style lang="scss" scoped>
.entries {
  &-text {
    margin: 0px 20% 0px 20%;
    font-size: 18px;
    line-height: 46px;
  }
  &-table {
    text-align: right;
    margin: 0px 10% 0px 10%;
    padding: 100px;
  }
}
</style>
