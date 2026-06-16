<template>
  <div class="user-menu-wrapper">
    <Avatar
      icon="pi pi-user"
      shape="circle"
      class="user-avatar"
      @click="toggleMenu"
      aria-haspopup="true"
      aria-controls="user-overlay-menu"
    />

    <Menu ref="menu" id="user-overlay-menu" :model="items" :popup="true" />

    <Dialog
      v-model:visible="infoVisible"
      header="Über Schematics2"
      :modal="true"
      :closable="true"
      style="width: 420px"
    >
      <p><strong>Schematics2</strong></p>
      <p>Es ermöglicht das Indexieren, Suchen und Verwalten von Schaltplan-Dokumenten und Effektbeschreibungen.</p>
      <p style="margin-top:1rem; font-size:0.85rem; color:#888">Versionen: App {{ APP_VERSION }}, Backend {{ info.version }}</p>
      <p>Status: {{ info.status }}</p>
    </Dialog>

    <Dialog
      v-model:visible="accountVisible"
      header="Mein Konto"
      :modal="true"
      :closable="true"
      style="width: 500px"
    >
      <div v-if="currentUser" style="display:grid; gap:1rem;">
        <div>
          <label style="font-weight:bold; color:#666;">E-Mail</label>
          <p>{{ currentUser.email }}</p>
        </div>
        <div>
          <label style="font-weight:bold; color:#666;">Name</label>
          <p>{{ currentUser.firstName }} {{ currentUser.lastName }}</p>
        </div>
        <div style="border-top:1px solid #e0e0e0; padding-top:1rem;">
          <label style="font-weight:bold; color:#666;">Adresse</label>
          <p v-if="currentUser.address">
            {{ currentUser.address.street }}<br />
            {{ currentUser.address.zipCode }} {{ currentUser.address.city }}
          </p>
          <p v-else style="color:#999;">Keine Adresse gespeichert</p>
        </div>
        <div style="border-top:1px solid #e0e0e0; padding-top:1rem; font-size:0.85rem; color:#888;">
          <p>Erstellt: {{ formatDate(currentUser.created) }}</p>
          <p>Zuletzt aktualisiert: {{ formatDate(currentUser.updated) }}</p>
        </div>
      </div>
      <div v-else style="text-align:center; padding:2rem; color:#999;">
        Daten werden geladen...
      </div>
    </Dialog>

    <Dialog
      v-model:visible="changePasswordVisible"
      header="Passwort ändern"
      :modal="true"
      :closable="true"
      style="width: 450px"
    >
      <div style="display:grid; gap:1rem;">
        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Aktuelles Passwort</label>
          <Password v-model="passwordForm.oldPassword" :toggleMask="true" style="width:100%;" />
        </div>
        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Neues Passwort</label>
          <Password v-model="passwordForm.newPassword" :toggleMask="true" style="width:100%;" />
          <small style="color:#999; display:block; margin-top:0.25rem;">Mindestens 8 Zeichen</small>
        </div>
        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Passwort wiederholen</label>
          <Password v-model="passwordForm.confirmPassword" :toggleMask="true" style="width:100%;" />
        </div>
        <div style="color:#e74c3c; font-size:0.9em;" v-if="passwordError">{{ passwordError }}</div>
        <div style="display:flex; gap:0.5rem; justify-content:flex-end; margin-top:1rem;">
          <Button label="Abbrechen" severity="secondary" @click="changePasswordVisible = false; resetPasswordForm()" />
          <Button label="Ändern" icon="pi pi-check" :loading="passwordChanging" @click="submitPasswordChange" />
        </div>
      </div>
    </Dialog>

    <Dialog
      v-model:visible="newUserVisible"
      header="Neuer Benutzer"
      :modal="true"
      :closable="true"
      style="width: 560px"
    >
      <div style="display:grid; gap:1rem;">
        <div style="display:grid; grid-template-columns:1fr 1fr; gap:0.75rem;">
          <div>
            <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Vorname</label>
            <InputText v-model="newUserForm.firstName" style="width:100%;" />
          </div>
          <div>
            <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Nachname</label>
            <InputText v-model="newUserForm.lastName" style="width:100%;" />
          </div>
        </div>

        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">E-Mail</label>
          <InputText v-model="newUserForm.email" type="email" style="width:100%;" />
        </div>

        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Passwort</label>
          <InputText v-model="newUserForm.password" style="width:100%;" />
          <small style="color:#999; display:block; margin-top:0.25rem;">Das Passwort ist vorbelegt und sichtbar, damit die manuelle Anlage schnell erfolgen kann.</small>
        </div>

        <div>
          <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Straße und Hausnummer</label>
          <InputText v-model="newUserForm.street" style="width:100%;" />
        </div>

        <div style="display:grid; grid-template-columns:1fr 1fr; gap:0.75rem;">
          <div>
            <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">PLZ</label>
            <InputText v-model="newUserForm.zipCode" style="width:100%;" />
          </div>
          <div>
            <label style="font-weight:bold; color:#666; display:block; margin-bottom:0.5rem;">Stadt</label>
            <InputText v-model="newUserForm.city" style="width:100%;" />
          </div>
        </div>

        <div style="color:#e74c3c; font-size:0.9em;" v-if="newUserError">{{ newUserError }}</div>

        <div style="display:flex; gap:0.5rem; justify-content:flex-end; margin-top:1rem;">
          <Button label="Abbrechen" severity="secondary" @click="newUserVisible = false" />
          <Button label="Speichern" icon="pi pi-check" :loading="newUserSaving" @click="submitNewUser" />
        </div>
      </div>
    </Dialog>

    <Dialog
      v-model:visible="userManagementVisible"
      header="Benutzerverwaltung"
      :modal="true"
      :closable="true"
      style="width: 1000px; max-height: 90vh"
    >
      <div style="display:flex; flex-direction:column; gap:1rem; height:100%;">
        <!-- Search bar and buttons -->
        <div style="display:flex; gap:0.5rem; align-items:center;">
          <span class="p-input-icon-left" style="flex:1;">
            <i class="pi pi-search" />
            <InputText
              v-model="searchUsername"
              placeholder="Nach Benutzer suchen..."
              style="width:100%;"
              @input="filterUsers"
            />
          </span>
          <Button
            :label="selectedUser?.deactive ? 'Aktivieren' : 'Deaktivieren'"
            :icon="selectedUser?.deactive ? 'pi pi-check' : 'pi pi-ban'"
            :severity="selectedUser?.deactive ? 'success' : 'danger'"
            :disabled="!selectedUser"
            @click="toggleDeactivateUser"
            :loading="deactivateLoading"
          />
          <Button
            label="Neuer Benutzer"
            icon="pi pi-user-plus"
            @click="openNewUserDialog"
          />
        </div>

        <!-- Users table -->
        <div style="flex:1; overflow-y:auto;">
          <DataTable
            :value="filteredUsers"
            selectionMode="single"
            v-model:selection="selectedUser"
            dataKey="id"
            :rows="10"
            stripedRows
            :paginator="filteredUsers.length > 10"
            @rowSelect="onRowSelect"
          >
            <Column field="email" header="E-Mail" style="width:25%;"></Column>
            <Column field="firstName" header="Vorname" style="width:15%;"></Column>
            <Column field="lastName" header="Nachname" style="width:15%;"></Column>
            <Column field="address.street" header="Adresse" style="width:25%;">
              <template #body="slotProps">
                <span v-if="slotProps.data.address">
                  {{ slotProps.data.address.street }},
                  {{ slotProps.data.address.zipCode }} {{ slotProps.data.address.city }}
                </span>
                <span v-else style="color:#999;">-</span>
              </template>
            </Column>
            <Column field="deactive" header="Status" style="width:12%;">
              <template #body="slotProps">
                <span v-if="slotProps.data.deactive" style="color:#e74c3c; font-weight:bold;">
                  Deaktiviert
                </span>
                <span v-else style="color:#27ae60; font-weight:bold;">
                  Aktiv
                </span>
              </template>
            </Column>
          </DataTable>
        </div>
      </div>
    </Dialog>
  </div>
</template> 
<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Avatar from 'primevue/avatar'
import Menu from 'primevue/menu'
import Dialog from 'primevue/dialog'
import Password from 'primevue/password'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'
import { APP_VERSION } from '../config'
import api from '../services/api'

const router = useRouter()
const { logout } = useAuth()
const { success, error: showError } = useToast()

const menu = ref(null)
const infoVisible = ref(false)
const accountVisible = ref(false)
const changePasswordVisible = ref(false)
const newUserVisible = ref(false)
const userManagementVisible = ref(false)
const info = ref({
  version: 'Loading...',
  status: 'Loading...',
})
const currentUser = ref(null)
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})
const passwordError = ref('')
const passwordChanging = ref(false)
const defaultAdminCreatedPassword = 'Start1234'
const newUserForm = ref({
  firstName: '',
  lastName: '',
  email: '',
  password: defaultAdminCreatedPassword,
  street: '',
  zipCode: '',
  city: '',
})
const newUserError = ref('')
const newUserSaving = ref(false)

// User Management
const allUsers = ref([])
const filteredUsers = ref([])
const selectedUser = ref(null)
const searchUsername = ref('')
const deactivateLoading = ref(false)

const isAdmin = computed(() => {
  const token = localStorage.getItem('schematics2_token')
  if (!token) {
    return false
  }

  const parts = token.split('.')
  if (parts.length < 2) {
    return false
  }

  try {
    const normalized = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
    const payload = JSON.parse(atob(padded))
    return Array.isArray(payload?.roles) && payload.roles.includes('admin')
  } catch {
    return false
  }
})

const items = computed(() => {
  const menuItems = [
    {
      label: 'Mein Konto',
      icon: 'pi pi-user',
      command: () => {
        fetchCurrentUser()
        accountVisible.value = true
      },
    },
    {
      label: 'Passwort ändern',
      icon: 'pi pi-key',
      command: () => {
        resetPasswordForm()
        changePasswordVisible.value = true
      },
    },
  ]

  if (isAdmin.value) {
    menuItems.push({
      label: 'Benutzerverwaltung',
      icon: 'pi pi-users',
      command: () => {
        fetchAllUsers()
        userManagementVisible.value = true
      },
    })
  }

  menuItems.push(
    {
      label: 'Info',
      icon: 'pi pi-info-circle',
      command: () => { infoVisible.value = true },
    },
    {
      label: 'Logout',
      icon: 'pi pi-sign-out',
      command: () => {
        logout()
        router.push('/')
      },
    },
  )

  return menuItems
})

function toggleMenu(event) {
  menu.value.toggle(event)
}

async function fetchBackendInfo() {
  try {
    const { data } = await api.get('/api/v1/info')
    info.value = {
      version: data.version || 'Unknown',
      status: data.status || 'Unknown',
    }
  } catch (err) {
    info.value = {
      version: 'Error',
      status: 'Error',
    }
  }
}

async function fetchCurrentUser() {
  try {
    const { data } = await api.get('/api/v1/users/me')
    currentUser.value = data
  } catch (err) {
    currentUser.value = null
  }
}

async function fetchAllUsers() {
  try {
    const { data } = await api.get('/api/v1/users')
    allUsers.value = data.users || []
    filteredUsers.value = allUsers.value
    selectedUser.value = null
    searchUsername.value = ''
  } catch (err) {
    showError('Fehler beim Laden der Benutzerliste')
    allUsers.value = []
    filteredUsers.value = []
  }
}

function filterUsers() {
  if (!searchUsername.value.trim()) {
    filteredUsers.value = allUsers.value
  } else {
    const query = searchUsername.value.toLowerCase()
    filteredUsers.value = allUsers.value.filter(user =>
      user.email.toLowerCase().includes(query) ||
      user.firstName.toLowerCase().includes(query) ||
      user.lastName.toLowerCase().includes(query)
    )
  }
  selectedUser.value = null
}

function onRowSelect(event) {
  selectedUser.value = event.data
}

function openNewUserDialog() {
  resetNewUserForm()
  userManagementVisible.value = false
  newUserVisible.value = true
}

async function toggleDeactivateUser() {
  if (!selectedUser.value) return

  deactivateLoading.value = true
  try {
    await api.patch(`/api/v1/users/${selectedUser.value.id}/deactivate`)
    success(selectedUser.value.deactive ? 'Benutzer aktiviert' : 'Benutzer deaktiviert')
    await fetchAllUsers()
  } catch (err) {
    showError('Fehler beim Ändern des Benutzerstatus')
  } finally {
    deactivateLoading.value = false
  }
}

function formatDate(timestamp) {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('de-DE')
}

function resetPasswordForm() {
  passwordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  passwordError.value = ''
}

function resetNewUserForm() {
  newUserForm.value = {
    firstName: '',
    lastName: '',
    email: '',
    password: defaultAdminCreatedPassword,
    street: '',
    zipCode: '',
    city: '',
  }
  newUserError.value = ''
}

function validatePasswordForm() {
  passwordError.value = ''
  
  if (!passwordForm.value.oldPassword) {
    passwordError.value = 'Aktuelles Passwort erforderlich'
    return false
  }
  if (!passwordForm.value.newPassword) {
    passwordError.value = 'Neues Passwort erforderlich'
    return false
  }
  if (passwordForm.value.newPassword.length < 8) {
    passwordError.value = 'Neues Passwort muss mindestens 8 Zeichen lang sein'
    return false
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordError.value = 'Passwörter stimmen nicht überein'
    return false
  }
  
  return true
}

async function submitPasswordChange() {
  if (!validatePasswordForm()) {
    return
  }
  
  passwordChanging.value = true
  try {
    await api.post('/api/v1/users/change-password', {
      oldPassword: passwordForm.value.oldPassword,
      newPassword: passwordForm.value.newPassword,
    })
    
    success('Passwort erfolgreich geändert')
    changePasswordVisible.value = false
    resetPasswordForm()
  } catch (err) {
    passwordError.value = err.response?.data?.message || 'Fehler beim Ändern des Passworts'
  } finally {
    passwordChanging.value = false
  }
}

function validateNewUserForm() {
  newUserError.value = ''

  const requiredFields = [
    newUserForm.value.firstName,
    newUserForm.value.lastName,
    newUserForm.value.email,
    newUserForm.value.password,
    newUserForm.value.street,
    newUserForm.value.zipCode,
    newUserForm.value.city,
  ]

  if (requiredFields.some((value) => !value || !value.trim())) {
    newUserError.value = 'Bitte alle Felder ausfüllen'
    return false
  }

  if (newUserForm.value.password.length < 8) {
    newUserError.value = 'Passwort muss mindestens 8 Zeichen lang sein'
    return false
  }

  return true
}

async function submitNewUser() {
  if (!validateNewUserForm()) {
    return
  }

  newUserSaving.value = true
  try {
    await api.post('/api/v1/auth/register', {
      firstName: newUserForm.value.firstName.trim(),
      lastName: newUserForm.value.lastName.trim(),
      email: newUserForm.value.email.trim(),
      password: newUserForm.value.password,
      street: newUserForm.value.street.trim(),
      zipCode: newUserForm.value.zipCode.trim(),
      city: newUserForm.value.city.trim(),
    })

    success('Benutzer wurde angelegt')
    newUserVisible.value = false
    // Wenn Benutzerverwaltung offen ist, neu laden
    if (userManagementVisible.value) {
      await fetchAllUsers()
      userManagementVisible.value = true
    }
  } catch (err) {
    newUserError.value = err.response?.data?.error || 'Fehler beim Anlegen des Benutzers'
    showError(newUserError.value)
  } finally {
    newUserSaving.value = false
  }
}

watch(infoVisible, (newVal) => {
  if (newVal) {
    fetchBackendInfo()
  }
})
</script>

<style scoped>
.user-avatar {
  cursor: pointer;
  background-color: var(--primary-color, #3b82f6);
  color: #fff;
  width: 2.4rem;
  height: 2.4rem;
  font-size: 1rem;
  flex-shrink: 0;
  transition: opacity 0.2s;
}

.user-avatar:hover {
  opacity: 0.85;
}

.user-menu-wrapper {
  display: flex;
  align-items: center;
}
</style>


