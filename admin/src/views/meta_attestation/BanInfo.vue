<script>
</script>
<template>
  <div>
    <CContainer>
      <CHeader class="mb-3 p-3">
        <h1>Attestation Info</h1>
      </CHeader>
    </CContainer>
    <CCol>
      <h2>platform ID -- > unique ID < -- > ban ID relationship</h2>

      <p>from unique ID, CANNOT retrieve platform ID with meta api</p>
      <p>CANNOT retrieve unique ID  directly by platform ID, this relationship is recorded when clients request attestation token with backend</p>
      <p>unique ID expire like 30 days, ban ID expire after duration expire</p>

      <br /><br />
      <CInput
          v-model="searchId"
          placeholder="Enter Meta Org-Scoped ID..."
          @keyup.enter="goToDetail"
      >
        <template #append>
          <CButton color="primary" @click="goToDetail">
            View platformId-uniqueId
          </CButton>
        </template>
      </CInput>

    </CCol>
  </div>
</template>

<script>
export default {
  data() {
    return {
      searchId: ''
    }
  },
  methods: {
    goToDetail() {
      if (!this.searchId.trim()) return;

      // Navigate to /meta-attestation/view/:platform_id
      this.$router.push({
        name: 'PlatformId-uniqueIdDetail',
        params: { platform_id: this.searchId.trim() }
      });
    }
  }
}
</script>