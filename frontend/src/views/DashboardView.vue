<template>
  <div class="dashboard-container">
    <el-header class="dashboard-header">
      <div class="header-logo">
        <span class="logo-text">Aussie EcoLens Workbench</span>
      </div>
      <div class="user-profile">
        <el-button type="danger" plain size="small" @click="handleLogout">Logout</el-button>
      </div>
    </el-header>

    <el-main class="dashboard-main">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card class="box-card upload-card">
            <template #header>
              <div class="card-header">
                <span class="title-text">Wildlife Observation Ingestion</span>
              </div>
            </template>
            
            <el-upload
              class="media-uploader"
              drag
              action="#"
              :auto-upload="true"
              :http-request="handleCustomUpload"
              :show-file-list="false"
              accept="image/*,video/*"
              :disabled="isUploading"
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                Drop wildlife media here or <em>click to upload</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  Supports image/video. Automatic cryptographic SHA-256 deduplication enforced.
                </div>
              </template>
            </el-upload>

            <div v-if="isUploading" class="status-overlay">
              <el-progress type="circle" :percentage="uploadPercentage" :status="uploadStatus" />
              <p class="status-tip-text">{{ statusMessage }}</p>
            </div>
          </el-card>
          <el-card class="box-card governance-card" style="margin-top: 20px;">
            <template #header>
              <div class="card-header"><span>Data Governance & Batch Management</span></div>
            </template>
            <p class="placeholder-text">Administrative batch destructive actions and classification controls will be implemented here.</p>
          </el-card>
        </el-col>

        <el-col :span="16">
          <el-card class="box-card search-card">
            <template #header>
              <div class="card-header"><span>Logical Search & Observation Grid</span></div>
            </template>

            <el-form :inline="true" :model="searchQuery" size="default" style="margin-bottom: -10px;">
              <el-form-item label="Species">
                <el-input v-model="searchQuery.species" placeholder="e.g., Alectura_lathami" clearable />
              </el-form-item>
              <el-form-item label="Tag">
                <el-input v-model="searchQuery.tag" placeholder="Filter by custom tag" clearable />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSearch">Search</el-button>
                <el-button @click="resetSearch">Reset</el-button>
              </el-form-item>
            </el-form>

            <el-divider style="margin: 15px 0;" />
            
            <el-table :data="observationList" v-loading="isTableLoading" style="width: 100%" border max-height="500">
              <el-table-column label="Thumbnail" width="120" align="center">
                <template #default="scope">
                  <el-image 
                    style="width: 50px; height: 50px; border-radius: 4px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                    :src="scope.row.thumbnail_display_url" 
                    :preview-src-list="[scope.row.file_url]"
                    preview-teleported
                    fit="cover"
                  >
                    <template #error>
                      <div style="display: flex; justify-content: center; align-items: center; width: 100%; height: 100%; background: #f5f7fa; color: #909399; font-size: 12px;">
                        No Thumb
                      </div>
                    </template>
                  </el-image>
                </template>
              </el-table-column>
              <el-table-column prop="primary_species" label="Inferred Species" width="240">
                <template #default="scope">
                  <el-tag :type="scope.row.primary_species === 'Alectura_lathami' ? 'success' : 'warning'" effect="dark">
                    {{ scope.row.primary_species || 'Analyzing...' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="Metadata Tags">
                <template #default="scope">
                  <el-tag v-for="tag in scope.row.tags" :key="tag" size="small" style="margin-right: 5px;" effect="plain">
                    {{ tag }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="Ingestion Date" width="160" />
              <el-table-column label="Actions" width="120" align="center">
                <template #default="scope">
                  <el-link 
                    :href="scope.row.file_download_url" 
                    target="_blank" 
                    type="primary" 
                    underline="never"
                  >
                    <el-button type="primary" size="small" link>Download</el-button>
                  </el-link>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </el-main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { UploadFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

const router = useRouter()
const isUploading = ref(false)
const uploadPercentage = ref(0)
const uploadStatus = ref('')
const statusMessage = ref('')

// Ingestion Core Configuration
const API_BASE_URL = 'https://aetsjr34k4.execute-api.us-east-1.amazonaws.com'


// Observation Grid Workspace Module
const isTableLoading = ref(false)
const observationList = ref([])

const QUERY_BASE_URL = 'https://aetsjr34k4.execute-api.us-east-1.amazonaws.com'

const searchQuery = ref({
  species: '',
  tag: ''
})

// Pagination tokens
const nextToken = ref('')
const hasMore = ref(false)

/**
 * Synchronizes layout display registry with backend observation store entries
 */
const fetchObservations = async () => {
  isTableLoading.value = true
  try {
    const token = localStorage.getItem('id_token')
    
    // Assemble query parameters based on active component state filters
    let queryParams = []
    if (searchQuery.value.species.trim()) {
      queryParams.push(`species=${encodeURIComponent(searchQuery.value.species.trim())}`)
    }
    if (searchQuery.value.tag.trim()) {
      queryParams.push(`tag=${encodeURIComponent(searchQuery.value.tag.trim())}`)
    }
    
    // Enforce API default pagination limit constraint
    queryParams.push('limit=10')

    if(nextToken.value) {
      queryParams.push(`next_token=${encodeURIComponent(nextToken.value)}`)
    }
    
    const queryString = queryParams.length ? `?${queryParams.join('&')}` : ''
    const finalUrl = `${QUERY_BASE_URL}/api/v1/observations${queryString}`
    
    console.log("Generated Target Query URL:", finalUrl)

    // Fetching real-time records
    const response = await axios.get(finalUrl, {
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if(response.data) {
      observationList.value = response.data.items || []
      nextToken.value = response.data.next_token || ''
      hasMore.value = response.data.has_more || false
    }
  
  } catch (error) {
    console.error("Failed to sync observation ledger:", error)
    ElMessage.error("Query dispatch failed. Verify API container status.")
    observationList.value = []
    nextToken.value = ''
    hasMore.value = false
  } finally {
    isTableLoading.value = false
  }
}

/**
 * Triggers the unified multi-conditional GET filter workflow
 */
const handleSearch = () => {
  nextToken.value = ''
  fetchObservations()
}

/**
 * Flushes active filter matrices and restores default observation stream
 */
const resetSearch = () => {
  searchQuery.value = { species: '', tag: '' }
  nextToken.value = ''
  fetchObservations()
}


// Fetch data when layout component mounts in viewport
onMounted(() => {
  fetchObservations()
})

/**
 * Generates a deterministic SHA-256 hash string from a file buffer.
 * Enforces client-side deduplication before storage ingress.
 */
const calculateSHA256 = async (file) => {
  const arrayBuffer = await file.arrayBuffer()
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}

/**
 * Custom interceptor for Element Plus upload lifecycle.
 * Manages checksum registry checks and direct streaming to the AWS S3 bucket.
 */
const handleCustomUpload = async (options) => {
  // Extract the actual file object passed by the Element Plus upload component
  const { file } = options
  isUploading.value = true
  uploadPercentage.value = 10
  uploadStatus.value = ''
  statusMessage.value = 'Analyzing asset signature...'

  try {
    // 1. Pre-process metadata classifications
    const fileType = file.type.startsWith('video/') ? 'video' : 'image'
    
    // 2. Generate SHA-256 Checksum on the Client Side
    const checksum = await calculateSHA256(file)
    uploadPercentage.value = 30
    statusMessage.value = 'Checking database deduplication registry...'

    // 3. Extract Cognito JWT Token for Secure Request Header
    const token = localStorage.getItem('id_token')

    // console.log({
    //   filename: file.name,
    //   content_type: file.type,
    //   size: file.size,
    //   checksum_sha256: checksum,
    //   file_type: fileType
    // });

    // 4. Check with the backend if this file already exists.
    // Send the file metadata and attach the Cognito token for authentication.
    const response = await axios.post(`${API_BASE_URL}/media/upload-url`, {
      filename: file.name,
      mime_type: file.type,
      size: file.size,
      checksum_sha256: checksum,
      file_type: fileType
    }, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    // Extract the specific fields from the backend's response based on API contract.
    const { duplicate, upload_url, file_id } = response.data

    if (duplicate) {
      // Skip the actual S3 upload to save time and just show a success message.
      uploadPercentage.value = 100
      uploadStatus.value = 'success'
      statusMessage.value = 'Deduplication success!'
      
      ElMessageBox.alert(
        `This file already exists in the archive (ID: ${file_id}).`,
        'Observation Deduplicated',
        { confirmButtonText: 'OK', type: 'success' }
      )
    } else {
      // Use the presigned URL from the backend to upload the actual file to S3.
      uploadPercentage.value = 60
      statusMessage.value = 'Uploading file directly to S3...'

      
      await axios.put(upload_url, file, {
        headers: { 'Content-Type': file.type }
      })

      uploadPercentage.value = 100
      uploadStatus.value = 'success'
      statusMessage.value = 'Upload completed!'
      ElMessage.success('New wildlife observation uploaded to cloud storage!')
    }
  } catch (error) {
    console.error('Ingestion lifecycle error:', error)
    uploadStatus.value = 'exception'
    statusMessage.value = 'Ingestion pipeline aborted.'
    ElMessage.error('Upload failed. Verify security tokens or API gateway availability.')
  } finally {
    // Delay resetting to let users see the 100% success state animation
    setTimeout(() => {
      isUploading.value = false
    }, 2000)
  }
}

/**
 * Handle user session termination
 */
const handleLogout = () => {
  localStorage.removeItem('id_token')
  ElMessage.info('Session invalidated successfully.')
  router.push('/login')
}
</script>

<style scoped>
.dashboard-container {
  min-height: 100vh;
  background-color: #f5f7fa;
}
.dashboard-header {
  background-color: #ffffff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}
.logo-text {
  font-size: 18px;
  font-weight: bold;
  color: #409EFF;
}
.dashboard-main {
  padding: 20px;
}
.upload-card {
  position: relative;
}
.title-text {
  font-weight: bold;
}
.media-uploader {
  margin-top: 10px;
}
.status-overlay {
  margin-top: 20px;
  text-align: center;
  background-color: #fafafa;
  padding: 15px;
  border-radius: 8px;
  border: 1px dashed #e0e0e0;
}
.status-tip-text {
  margin-top: 10px;
  font-size: 13px;
  color: #606266;
}
.placeholder-text {
  color: #909399;
  font-style: italic;
  font-size: 13px;
}
</style>