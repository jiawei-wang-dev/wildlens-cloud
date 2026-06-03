<template>
  <div class="login-container">
    <el-card class="login-card" shadow="always">
      <div class="login-header">
        <el-icon class="logo-icon" :size="40" color="#409EFF"><CameraFilled /></el-icon>
        <h2>Aussie EcoLens</h2>
        <p class="subtitle">Multi-Cloud Serverless Wildlife Observation Platform</p>
      </div>

      <div class="login-body">
        <!-- Displayed dynamically when parsing authentication codes from AWS -->
        <el-alert
          v-if="isProcessing"
          title="Authenticating with AWS Cognito..."
          type="info"
          :closable="false"
          show-icon
        />
        
        <!-- Standard interaction zone before redirection -->
        <div v-else class="action-zone">
          <p class="tip-text">Secure access powered by AWS Cognito. Please sign in or register to manage wildlife observations.</p>
          <el-button type="primary" size="large" class="login-btn" @click="redirectToCognito">
            Sign In / Sign Up via AWS
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CameraFilled } from '@element-plus/icons-vue' 

const router = useRouter()
const isProcessing = ref(false)

// Integrated with the real AWS Cognito URL parameters provided by your teammate!
const COGNITO_HOSTED_UI_URL = 'https://us-east-1zhpmn5rx5.auth.us-east-1.amazoncognito.com/login?client_id=4vaujh7rjc3as3q38pqlva6iet&response_type=code&scope=email+openid+phone&redirect_uri=http://localhost:3000/login'

// Dispatches the browser viewport directly to AWS secure login dashboard
const redirectToCognito = () => {
  window.location.href = COGNITO_HOSTED_UI_URL
}

// Lifecycle Hook: Catches response codes from AWS once redirected back to our Vite local server
onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const code = urlParams.get('code')
  
  if (code) {
    isProcessing.value = true
    
    // Save the authentication receipt locally to grant passage through the Router Guard
    localStorage.setItem('id_token', 'verified_auth_code_' + code)
    ElMessage.success('Authenticated via AWS Cognito successfully!')
    
    // Smoothly transition into the protected application dashboard
    router.push('/dashboard')
  }
})
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #2b5876 0%, #4e4376 100%);
}
.login-card {
  width: 450px;
  border-radius: 12px;
  padding: 20px;
}
.login-header {
  text-align: center;
  margin-bottom: 30px;
}
.logo-icon {
  margin-bottom: 10px;
}
.login-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}
.subtitle {
  font-size: 13px;
  color: #909399;
  margin-top: 8px;
  text-align: center;
}
.tip-text {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  text-align: center;
  margin-bottom: 25px;
}
.login-btn {
  width: 100%;
  font-weight: bold;
}
</style>