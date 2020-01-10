<template>
  <div class="app-container">

      <el-form ref="dataForm" :model="config" label-position="left" label-width="200px" style="width: 600px; margin-left:50px;">
        <el-form-item label="Log Level">
          <el-select v-model="config.log_level" class="filter-item" placeholder="Please select">
            <el-option v-for="item in log_level_options" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="Replicate Timeout">
          <el-input v-model="config.replicate_timeout" placeholder="1000" />
        </el-form-item>
        <el-form-item label="Replicate Unit Size">
          <el-input v-model="config.replicate_unit_size" placeholder="1024" />
        </el-form-item>
        <el-form-item label="Max Log Size">
          <el-input v-model="config.max_log_size" placeholder="1048576" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleUpdateConfig()">
            Update
          </el-button>
        </el-form-item>
      </el-form>
  </div>
</template>
<script>

import { GetConfig, UpdateConfig } from '@/api/config'

export default {
  name: 'Config',
  data() {
    return {
      config: {
        log_level: '',
        replicate_timeout: 0,
        replicate_unit_size: 0,
        max_log_size: 0
      },
      log_level_options: ['DEBUG', 'INFO', 'WARNING', 'ERROR', 'PANIC', 'FATAL']
    }
  },
  created() {
    this.fetchData()
  },
  methods: {
    fetchData() {
      GetConfig().then(response => {
        if (response != null) {
          this.config = response
        }
      })
    },
    tableHeaderColor({ row, column, rowIndex, columnIndex }) {
      if (rowIndex === 0) {
        return 'background-color: gray; color: #fff; font-weight: 500;'
      }
    },
    handleUpdateConfig() {
      const param = {
        log_level: this.config.log_level,
        replicate_timeout: parseInt(this.config.replicate_timeout),
        replicate_unit_size: parseInt(this.config.replicate_unit_size),
        max_log_size: parseInt(this.config.max_log_size)
      }
      UpdateConfig(param).then(() => {
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
