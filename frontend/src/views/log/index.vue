<template>
  <div class="app-container">

    <div class="post-container">
      <el-form>
        <el-row>
          <el-col :span="8">
            <el-form-item label-width="100px" label="Request ID:">
              <el-input v-model="log.request_id" placeholder="request id" style="width: 400px" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label-width="100px" label="Command:">
              <el-input v-model="log.command" placeholder="command" style="width: 400px" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-button type="primary" @click="handleAddLog()">
              Add
            </el-button>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <div class="entries-table">
      <el-table :data="logs.entries" border fit highlight-current-row>
        <el-table-column align="center" prop="request_id" label="Request ID" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="term" label="Term" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="command" label="Command" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="append_time" label="Append Time" min-width="15%">
        </el-table-column>
        <el-table-column align="center" prop="apply_time" label="Apply Time" min-width="15%">
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>
<script>

import { GetLogs, AddLog } from '@/api/log'

export default {
  name: 'Log',
  timer: '',
  data() {
    return {
      logs: {
        total: 0,
        entries: []
      },
      log: {
        request_id: '',
        command: ''
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
      GetLogs().then(response => {
        if (response != null) {
          this.logs = response
        }
      })
    },
    handleAddLog() {
      AddLog(this.log).then(() => {
        this.$notify({
          title: 'Success',
          message: 'Successfully',
          type: 'success',
          duration: 2000
        })
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.post-container {
  position: relative;
  margin: 0px 15% 0px 15%;
}
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
.postInfo-container {
  position: relative;
  margin-bottom: 10px;

  .postInfo-container-item {
    float: left;
  }
}
</style>
