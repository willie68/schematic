<template>
  <section class="card" style="max-width: 520px; margin: 0 auto;">
    <h2 v-if="!showForcePasswordChange">Anmeldung</h2>
    <h2 v-else>Passwort ändern erforderlich</h2>

    <!-- Login Form -->
    <div v-if="!showForcePasswordChange" style="display:grid; gap:1.5rem;">
      <div style="box-sizing:border-box; width:100%;">
        <label for="user" style="display:block; font-weight:500; font-size:0.95rem; margin-bottom:0.5rem;">Benutzername</label>
        <InputText id="user" v-model="username" @keyup.enter="login" style="width:100%; min-height:42px; box-sizing:border-box;" />
      </div>

      <div style="box-sizing:border-box; width:100%;">
        <label for="pass" style="display:block; font-weight:500; font-size:0.95rem; margin-bottom:0.5rem;">Passwort</label>
        <Password id="pass" v-model="password" :feedback="false" toggleMask @keyup.enter="login" style="width:100%; min-height:42px; box-sizing:border-box;" />
      </div>

      <Button label="Einloggen" icon="pi pi-sign-in" @click="login" />
    </div>

    <!-- Force Password Change Form -->
    <div v-else style="display:grid; gap:1.5rem;">
      <div style="background:#fff3cd; border:1px solid #ffc107; border-radius:4px; padding:1rem; color:#856404; font-size:0.9rem;">
        <i class="pi pi-exclamation-triangle" style="margin-right:0.5rem;"></i>
        Sie müssen Ihr Passwort beim ersten Login ändern.
      </div>

      <div>
        <label for="newPass" style="display:block; font-weight:500; font-size:0.95rem; margin-bottom:0.5rem;">Neues Passwort</label>
        <Password id="newPass" v-model="forcePasswordForm.newPassword" :feedback="false" toggleMask style="width:100%;" />
        <small style="color:#999; display:block; margin-top:0.25rem;">Mindestens 8 Zeichen</small>
      </div>

      <div>
        <label for="confirmPass" style="display:block; font-weight:500; font-size:0.95rem; margin-bottom:0.5rem;">Passwort wiederholen</label>
        <Password id="confirmPass" v-model="forcePasswordForm.confirmPassword" :feedback="false" toggleMask style="width:100%;" />
      </div>

      <Button label="Passwort ändern" icon="pi pi-check" @click="submitForcedPasswordChange" :loading="isChangingPassword" />
    </div>

    <Message v-if="message" :severity="messageType" style="margin-top:1rem">{{ message }}</Message>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import Message from 'primevue/message'
import api from '../services/api'
import { useAuth } from '../composables/useAuth'

const router = useRouter()
const { setLoggedIn } = useAuth()
const username = ref('')
const password = ref('')
const message = ref('')
const messageType = ref('info')
const showForcePasswordChange = ref(false)
const isChangingPassword = ref(false)
const forcePasswordForm = ref({
  newPassword: '',
  confirmPassword: '',
})
const currentUserEmail = ref('')

async function login() {
  message.value = ''
  try {
    const { data } = await api.post('/api/v1/auth/login', {
      username: username.value,
      password: password.value,
    })
    
    localStorage.setItem('schematics2_token', data.token)
    setLoggedIn(true)
    
    // Check if user must change password
    if (data.changePassword) {
      currentUserEmail.value = data.email
      showForcePasswordChange.value = true
      message.value = 'Bitte ändern Sie Ihr Passwort, um fortzufahren.'
      messageType.value = 'warn'
      return
    }
    
    messageType.value = 'success'
    message.value = 'Login erfolgreich.'
    setTimeout(() => {
      router.push('/')
    }, 500)
  } catch (err) {
    messageType.value = 'error'
    message.value = err?.response?.data?.error || 'Login fehlgeschlagen.'
  }
}

async function submitForcedPasswordChange() {
  message.value = ''
  
  // Validate
  if (!forcePasswordForm.value.newPassword) {
    message.value = 'Neues Passwort erforderlich'
    messageType.value = 'error'
    return
  }
  
  if (forcePasswordForm.value.newPassword.length < 8) {
    message.value = 'Passwort muss mindestens 8 Zeichen lang sein'
    messageType.value = 'error'
    return
  }
  
  if (forcePasswordForm.value.newPassword !== forcePasswordForm.value.confirmPassword) {
    message.value = 'Passwörter stimmen nicht überein'
    messageType.value = 'error'
    return
  }
  
  isChangingPassword.value = true
  try {
    await api.post('/api/v1/users/change-password', {
      oldPassword: password.value,
      newPassword: forcePasswordForm.value.newPassword,
    })
    
    messageType.value = 'success'
    message.value = 'Passwort erfolgreich geändert. Weitergeleitet...'
    
    setTimeout(() => {
      router.push('/')
    }, 1500)
  } catch (err) {
    messageType.value = 'error'
    message.value = err?.response?.data?.message || 'Fehler beim Ändern des Passworts'
  } finally {
    isChangingPassword.value = false
  }
}
</script>

<style scoped>
:deep(.p-inputtext),
:deep(.p-password) {
  width: 100% !important;
  box-sizing: border-box !important;
}

:deep(.p-password .p-password-input) {
  width: 100% !important;
  box-sizing: border-box !important;
}

:deep(.p-password) {
  display: flex !important;
  width: 100% !important;
}
</style>

