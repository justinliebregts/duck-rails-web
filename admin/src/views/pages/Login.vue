<template>
<div class="c-app align-items-center justify-content-center p-3"><form class="card p-4 login-card" @submit.prevent="login">
<h1>Duck Rails Admin</h1><p>Enter the API admin token to access reports and administration tools.</p>
<label for="admin-token">Admin token</label><input id="admin-token" v-model="token" class="form-control mb-3" type="password" autocomplete="current-password" required :disabled="busy">
<p v-if="error" role="alert" class="alert alert-danger">{{ error }}</p><button type="submit" class="btn btn-primary" :disabled="busy || !token.trim()">{{ busy ? 'Checking…' : 'Sign in' }}</button>
</form></div></template>
<script>
export default {
name: 'Login', data: () => ({ token: '', busy: false, error: '' }),
methods: { async login() {
this.busy = true; this.error = ''
const token = this.token.trim().replace(/^Bearer\s+/i, '')
try {
await this.axios.get(this.$gstate.defaultUrl + '/diagnostics/admin/reports', { headers: { Authorization: 'Bearer ' + token }, timeout: 20000 })
this.$auth.tmpLoginForever('Bearer ' + token); this.token = ''; this.$router.replace('/diagnostics')
} catch (error) { this.error = error.response && [401, 403].includes(error.response.status) ? 'The admin token was rejected.' : 'Could not connect to the admin API. Please try again.' }
finally { this.busy = false }
} }
}
</script>
<style scoped>.login-card { width: 100%; max-width: 480px; }</style>
