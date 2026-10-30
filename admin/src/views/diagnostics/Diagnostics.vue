<template>
  <div>
    <h1>Diagnostics</h1><p class="text-muted">Reports from testers and automatic game checks.</p>
    <div class="card p-3 mb-3">
      <div class="d-flex align-items-center controls">
        <label for="status">Status</label>
        <select id="status" v-model="status" class="form-control w-auto" :disabled="loading || saving" @change="offset = 0; load()"><option value="open">Open</option><option value="resolved">Resolved</option><option value="all">All</option></select>
        <button class="btn btn-primary" :disabled="loading || saving" @click="load">{{ loading ? 'Loading…' : 'Refresh' }}</button>
        <span v-if="!error && !loading">{{ total }} reports</span>
      </div>
      <p v-if="error" role="alert" class="alert alert-danger mt-3">{{ error }}</p>
      <p v-if="notice" role="status" class="alert alert-success mt-3">{{ notice }}</p>
      <div class="table-responsive mt-3"><table class="table table-hover">
        <thead><tr><th>Created</th><th>Summary</th><th>Category</th><th>Platform</th><th>Status</th><th></th></tr></thead>
        <tbody><tr v-for="report in reports" :key="report.id"><td>{{ date(report.created) }}</td><td class="summary">{{ report.summary }}</td><td>{{ report.category }}</td><td>{{ report.platform }}</td><td>{{ report.status }}</td><td><button class="btn btn-sm btn-outline-primary" :disabled="detailLoading || saving" @click="open(report.id)">View</button></td></tr></tbody>
      </table></div>
      <p v-if="!loading && !error && !reports.length">No {{ status === 'all' ? '' : status }} reports.</p>
      <div class="d-flex controls"><button class="btn btn-outline-secondary" :disabled="offset === 0 || loading || saving" @click="offset -= 100; load()">Previous</button><span>Page {{ Math.floor(offset / 100) + 1 }}</span><button class="btn btn-outline-secondary" :disabled="offset + 100 >= total || loading || saving" @click="offset += 100; load()">Next</button></div>
    </div>
    <p v-if="detailLoading" role="status">Loading report…</p>
    <section v-if="selected" class="card p-3" aria-label="Report details">
      <div class="d-flex justify-content-between"><h2>Report details</h2><button class="btn btn-outline-secondary" :disabled="saving" @click="selected = null">Close</button></div>
      <h3 class="h5 mt-3">{{ selected.summary }}</h3><p>{{ selected.id }} · {{ date(selected.created) }} · {{ selected.status }}</p>
      <p class="summary">Build: {{ selected.build }} · Platform: {{ selected.platform }} · Category: {{ selected.category }}</p>
      <p v-if="selected.resolution"><strong>Resolution:</strong> {{ selected.resolution }}</p>
      <p v-if="payloadError" class="alert alert-warning">The payload is not valid JSON. Download is still available.</p>
      <template v-if="payload">
        <details open><summary>Game snapshot</summary><pre>{{ pretty(payload.snapshot) }}</pre></details>
        <details><summary>History ({{ (payload.history || []).length }} samples)</summary><pre>{{ pretty(payload.history) }}</pre></details>
        <details><summary>Events ({{ (payload.events || []).length }})</summary><pre>{{ pretty(payload.events) }}</pre></details>
        <details><summary>Test preset</summary><pre>{{ pretty(payload.preset) }}</pre></details>
      </template>
      <button class="btn btn-outline-primary align-self-start my-3" @click="download">Download report JSON</button>
      <label for="resolution">Fix and verification notes (required to resolve)</label><textarea id="resolution" v-model="resolution" class="form-control mb-3" rows="3" maxlength="1000" :disabled="saving"></textarea>
      <button class="btn btn-primary align-self-start" :disabled="saving || (selected.status === 'open' && !resolution.trim())" @click="changeStatus">{{ saving ? 'Saving…' : selected.status === 'open' ? 'Resolve report' : 'Reopen report' }}</button>
    </section>
  </div>
</template>
<script>
export default {
  name: 'Diagnostics',
  data: () => ({ status: 'open', offset: 0, total: 0, reports: [], selected: null, payload: null, payloadError: false, resolution: '', loading: false, detailLoading: false, saving: false, error: '', notice: '' }),
  computed: { endpoint() { return `${this.$gstate.defaultUrl}/diagnostics/admin/reports` } },
  mounted() { this.load() },
  methods: {
    date(value) { return value ? new Date(value).toLocaleString() : 'Unknown' },
    pretty(value) { return JSON.stringify(value == null ? null : value, null, 2) },
    failure(error) { return error.response && [401, 403].includes(error.response.status) ? 'Admin access rejected. Sign out and sign in with the configured admin token.' : 'Request failed. Check the API connection and try again.' },
    async load() {
      this.loading = true; this.error = ''
      try {
        const response = await this.axios.get(this.endpoint, { ...this.$auth.apiOptions, timeout: 20000, params: { status: this.status, offset: this.offset } })
        this.reports = response.data.data.reports || []; this.total = response.data.data.total
        if (this.offset > 0 && this.offset >= this.total) { this.offset = Math.max(0, Math.floor((this.total - 1) / 100) * 100); await this.load() }
      } catch (error) { this.error = this.failure(error); this.reports = [] }
      finally { this.loading = false }
    },
    async open(id) {
      this.detailLoading = true; this.selected = null; this.error = ''; this.notice = ''; this.payload = null; this.payloadError = false
      try {
        const response = await this.axios.get(`${this.endpoint}/${id}`, { ...this.$auth.apiOptions, timeout: 20000 })
        this.selected = response.data.data; this.resolution = this.selected.resolution || ''
        try { this.payload = JSON.parse(this.selected.payload); if (!this.payload || typeof this.payload !== 'object') this.payloadError = true }
        catch (_) { this.payloadError = true }
      } catch (error) { this.error = this.failure(error) }
      finally { this.detailLoading = false }
    },
    async changeStatus() {
      if (!this.selected || this.saving) return
      const status = this.selected.status === 'open' ? 'resolved' : 'open'
      if (status === 'resolved' && !this.resolution.trim()) return
      this.saving = true; this.error = ''; this.notice = ''
      try {
        await this.axios.post(`${this.endpoint}/${this.selected.id}/status`, { status, resolution: this.resolution.trim() }, { ...this.$auth.apiJsonOptions, timeout: 20000 })
        this.selected = { ...this.selected, status, resolution: this.resolution.trim() }; this.notice = status === 'resolved' ? 'Report resolved.' : 'Report reopened.'; await this.load()
      } catch (error) { this.error = this.failure(error) }
      finally { this.saving = false }
    },
    download() {
      const url = URL.createObjectURL(new Blob([this.selected.payload], { type: 'application/json' }))
      const link = document.createElement('a'); link.href = url; link.download = `${this.selected.id}.json`; link.click(); setTimeout(() => URL.revokeObjectURL(url), 1000)
    }
  }
}
</script>
<style scoped>
.controls { gap: 12px; flex-wrap: wrap; align-items: center; }
.summary { overflow-wrap: anywhere; max-width: 500px; }
details { margin-top: 12px; } summary { cursor: pointer; font-weight: 600; }
pre { max-height: 360px; overflow: auto; background: #f3f5f7; padding: 16px; white-space: pre-wrap; overflow-wrap: anywhere; margin-top: 8px; }
</style>
