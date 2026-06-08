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
                <span class="title-text">Upload Wildlife Media</span>
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
              <div class="card-header"><span>Batch Management & Alerts</span></div>
            </template>
            <div class="governance-section">
              <h4 class="section-title">Batch Record Actions</h4>
              <p class="section-desc">Selected: <strong>{{ selectedRows.length }}</strong> items</p>
              <el-input v-model="bulkTagInput" placeholder="Type custom tag name..." size="default" style="margin-bottom: 12px;" clearable />
              <div class="btn-group">
                <el-button type="primary" @click="handleBulkTag(1)" :disabled="!selectedRows.length">Add Tag</el-button>
                <el-button type="warning" @click="handleBulkTag(0)" :disabled="!selectedRows.length">Remove Tag</el-button>
                <el-button type="danger" @click="handleBulkDelete" :disabled="!selectedRows.length">Delete Files</el-button>
              </div>
            </div>
            <el-divider style="margin: 20px 0;" />
            <div class="governance-section">
              <h4 class="section-title">AWS SNS Wildlife Alerts</h4>
              <el-input v-model="notificationForm.email" placeholder="your-email@example.com" size="default" style="margin-bottom: 10px;" />
              <el-input v-model="notificationForm.tag" placeholder="Species tag to subscribe (e.g., dingo)" size="default" style="margin-bottom: 12px;" />
              <div class="btn-group">
                <el-button type="success" @click="handleSubscribe(1)">Subscribe</el-button>
                <el-button type="info" plain @click="handleSubscribe(0)">Unsubscribe</el-button>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :span="16">
          <el-card class="box-card search-card">
            <template #header>
              <div class="card-header"><span>Search & Gallery</span></div>
            </template>

            <el-form :inline="true" :model="searchQuery" size="default" style="margin-bottom: -10px;">
              <el-form-item label="Species">
                <el-input v-model="searchQuery.species" placeholder="e.g., Alectura_lathami" clearable />
              </el-form-item>
              <el-form-item label="Tags">
                <el-input v-model="searchQuery.tag" placeholder="e.g., wild cute (separated by spaces)" style="width: 260px;" clearable />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSearch">Search</el-button>
                <el-button @click="resetSearch">Reset</el-button>
              </el-form-item>
            </el-form>

            <div style="margin-top: 15px; display: flex; gap: 15px; align-items: center; background: #fafafa; padding: 12px; border-radius: 4px; border: 1px dashed #e0e0e0;">
              <div style="display: flex; align-items: center; gap: 6px;">
                <span style="font-size: 12px; color: #606266;">Advanced JSON:</span>
                <el-input v-model="jsonTagQueryStr" placeholder='e.g., {"koala":3}' size="small" style="width: 150px;" clearable />
                <el-button type="primary" size="small" @click="handleJsonTagQuery">JSON Qry</el-button>
              </div>
              <div style="display: flex; align-items: center; gap: 6px; border-left: 1px solid #dcdfe6; padding-left: 15px;">
                <el-input v-model="reverseQueryUrl" placeholder="Paste thumbnail URL..." size="small" style="width: 180px;" clearable />
                <el-button type="warning" size="small" @click="handleThumbnailLookup">Reverse Lookup</el-button>
              </div>
              <div style="border-left: 1px solid #dcdfe6; padding-left: 15px;">
                <el-upload action="#" :auto-upload="true" :http-request="handleImageBasedSearch" :show-file-list="false" accept="image/*">
                  <el-button type="success" size="small">Query By Image File</el-button>
                </el-upload>
              </div>
            </div>
            <el-divider style="margin: 15px 0;" />
            
            <el-table :data="observationList" v-loading="isTableLoading" style="width: 100%" border max-height="500" @selection-change="handleSelectionChange">
              <el-table-column type="selection" width="50" align="center" />
              <el-table-column label="Thumbnail" width="120" align="center">
                <template #default="scope">
                  <el-image 
                    style="width: 50px; height: 50px; border-radius: 4px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                    :src="scope.row.thumbnail_display_url || scope.row.thumbnail_url" 
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
              <el-table-column label="Ingestion Date" width="190">
                <template #default="scope">
                  {{ formatMelbourneTime(scope.row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column label="Actions" width="120" align="center">
                <template #default="scope">
                  <el-link 
                    :href="scope.row.file_download_url || scope.row.file_url" 
                    target="_blank" 
                    type="primary" 
                    underline="never"
                  >
                    <el-button type="primary" size="small" link>Download</el-button>
                  </el-link>
                </template>
              </el-table-column>
            </el-table>
            <div class="pagination-footer" style="margin-top: 15px; display: flex; justify-content: space-between; align-items: center;">
              <span class="placeholder-text" style="font-size: 13px; color: #909399;">
                {{ nextToken ? 'More wildlife historical captures pending authorization...' : 'End of synchronization ledger reached.' }}
              </span>
              <div class="page-buttons">
                <button 
                  class="el-button el-button--primary" 
                  :disabled="!hasMore" 
                  @click="handleNextPage"
                  style="padding: 8px 15px; font-size: 12px; border-radius: 4px;"
                >
                  Next Page
                </button>
              </div>
            </div>
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

const jsonTagQueryStr = ref('')
const reverseQueryUrl = ref('')

// Pagination tokens
const nextToken = ref('')
const hasMore = ref(false)

const selectedRows = ref([])
const bulkTagInput = ref('')
const notificationForm = ref({
  email: '',
  tag: ''
})

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
      const tags = searchQuery.value.tag.trim().split(/\s+/).filter(Boolean)
      tags.forEach(t => {
        queryParams.push(`tag=${encodeURIComponent(t)}`)
      })
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
 * Evaluates pagination state and advances the grid registry to the subsequent cursor block
 */
const handleNextPage = () => {
  if (hasMore.value && nextToken.value) {
    fetchObservations()
  }
}

/**
 * Flushes active filter matrices and restores default observation stream
 */
const resetSearch = () => {
  searchQuery.value = { species: '', tag: '' }
  jsonTagQueryStr.value = ''
  reverseQueryUrl.value = ''
  nextToken.value = ''
  fetchObservations()
}

/**
 * 1. Advanced Tag Query with Count Constraints (POST /api/v1/query/tags)
 * Evaluates composite JSON objects for minimum fauna element count distribution filters.
 */
const handleJsonTagQuery = async () => {
  if (!jsonTagQueryStr.value.trim()) return ElMessage.warning('Please enter a JSON object first.')
  try {
    isTableLoading.value = true
    const token = localStorage.getItem('id_token')
    const parsedJson = JSON.parse(jsonTagQueryStr.value.trim())

    const response = await axios.post(`${QUERY_BASE_URL}/api/v1/query/tags`, parsedJson, {
      headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }
    })
    if (response.data) {
      observationList.value = response.data.files || []
      nextToken.value = ''
      hasMore.value = false
    }
  } catch (e) {
    console.error('Advanced JSON query error:', e)
    ElMessage.error('Invalid JSON format or API error.')
    observationList.value = []
    nextToken.value = ''
    hasMore.value = false
  } finally { isTableLoading.value = false }
}

/**
 * 2. Reverse Thumbnail Mapping Lookup (GET /api/v1/observations/lookup?thumbnail_url=...)
 * Resolves full-size source media locations via an explicit thumbnail resource pointer tracking request.
 */
const handleThumbnailLookup = async () => {
  if (!reverseQueryUrl.value.trim()) return ElMessage.warning('Please enter a URL first.')
  try {
    isTableLoading.value = true
    const token = localStorage.getItem('id_token')
    const response = await axios.get(`${QUERY_BASE_URL}/api/v1/observations/lookup?thumbnail_url=${encodeURIComponent(reverseQueryUrl.value.trim())}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (response.data && response.data.file_url) {
      ElMessageBox.alert(`<strong>Full Image Asset Found:</strong><br><a href="${response.data.file_url}" target="_blank">${response.data.file_url}</a>`, 'Lookup Success', { confirmButtonText: 'Open', dangerouslyUseHTMLString: true, type: 'success' }).then(() => {
        window.open(response.data.file_url, '_blank')
      })
    } else { ElMessage.info('No match found.') }
  } catch (e) { ElMessage.error('Lookup rejected by Gateway.') }
  finally { isTableLoading.value = false }
}

/**
 * 3. Transient Content Search via Image Upload (POST /api/v1/observations/search-by-file)
 * Dispatches a temporary media binary stream for inference evaluation to query stored assets without storage persistence.
 */
const handleImageBasedSearch = async (options) => {
  try {
    isTableLoading.value = true
    const token = localStorage.getItem('id_token')
    const formData = new FormData()
    formData.append('file', options.file)

    const response = await axios.post(`${QUERY_BASE_URL}/api/v1/observations/search-by-file`, formData, {
      headers: { 'Authorization': `Bearer ${token}`}
    })
    if (response.data) {
      observationList.value = response.data.items || response.data || []
      ElMessage.success('Transient query search finished!')
    }
  } catch (e) { ElMessage.error('Image search aborted.') }
  finally { isTableLoading.value = false }
}

// Fetch data when layout component mounts in viewport
onMounted(() => {
  fetchObservations()
})

const handleSelectionChange = (val) => {
  selectedRows.value = val
}

const handleBulkTag = async (operationType) => {
  if (!bulkTagInput.value.trim()) {
    return ElMessage.warning('Please enter a valid tag name first.')
  }
  const token = localStorage.getItem('id_token')
  const fileUrls = selectedRows.value.map(row => row.file_url)
  const payload = {
    urls: fileUrls,
    tags: [bulkTagInput.value.trim()],
    operation: operationType
  }
  try {
    isTableLoading.value = true
    await axios.post(`${QUERY_BASE_URL}/api/v1/tags/update`, payload, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    ElMessage.success('Bulk tag operation successfully completed!')
    bulkTagInput.value = ''
    await fetchObservations()
  } catch (error) {
    console.error('Bulk tagging error:', error)
    ElMessage.error('Failed to update tags. Verify gateway parameters.')
  } finally {
    isTableLoading.value = false
  }
}

const handleBulkDelete = () => {
  ElMessageBox.confirm(
    `Are you sure you want to completely erase these ${selectedRows.value.length} records from both S3 buckets and DynamoDB tables? This action is destructive and permanent.`,
    'Warning: Cloud Asset Destruction Request',
    {
      confirmButtonText: 'Confirm Purge',
      cancelButtonText: 'Abort',
      type: 'danger'
    }
  ).then(async () => {
    const token = localStorage.getItem('id_token')
    const fileUrls = selectedRows.value.map(row => row.file_url)
    try {
      isTableLoading.value = true
      await axios.delete(`${QUERY_BASE_URL}/api/v1/files`, {
        headers: { 'Authorization': `Bearer ${token}` },
        data: {urls: fileUrls}
      })
      ElMessage.success('Target assets successfully removed from cloud registry.')
      await fetchObservations()
    } catch (error) {
      console.error('Purge transaction error:', error)
      ElMessage.error('Purge action aborted by remote gateway.')
    } finally {
      isTableLoading.value = false
    }
  }).catch(() => {
    ElMessage.info('Purge operation cancelled.')
  })
}

const handleSubscribe = async (actionType) => {
  const { email, tag } = notificationForm.value
  if (!email.trim() || !tag.trim()) {
    return ElMessage.warning('Both Email destination and target Tag matching fields are mandatory.')
  }
  const token = localStorage.getItem('id_token')
  const payload = {
    email: email.trim(),
    tag_name: tag.trim(),
    action: actionType === 1 ? 'subscribe' : 'unsubscribe'
  }
  try {
    await axios.post(`${API_BASE_URL}/notifications/subscribe`, payload, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (actionType === 1) {
      ElMessage.success('Subscription pipeline deployed! Verify confirmation link via AWS email receipt.')
    } else {
      ElMessage.success('Target alert endpoint unlinked successfully.')
    }
    notificationForm.value = { email: '', tag: '' }
  } catch (error) {
    console.error('SNS pipeline registration error:', error)
    const serverMessage = error.response?.data || 'Notification dispatch registry failed.'
    ElMessage.error(serverMessage)
  }
}

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

/**
 * Formats an ISO/UTC RFC3339 timestamp string into localized Australia/Melbourne time.
 * Automatically accounts for daylight saving time (AEST/AEDT) offsets.
 */
const formatMelbourneTime = (value) => {
  if (!value) return ''

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('en-AU', {
    timeZone: 'Australia/Melbourne',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }).format(date)
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

.governance-section {
  text-align: left;
}
.section-title {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #303133;
}
.section-desc {
  font-size: 12px;
  color: #909399;
  margin-bottom: 10px;
}
.btn-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.card-header span {
  font-weight: bold;
}
</style>