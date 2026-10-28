<template>
  <CCard class="mb-4">
    <CCardHeader>
      <h3>Local Attestation Records</h3>
      <h6>(Click to select Unique ID)</h6>
    </CCardHeader>
    <CDataTable
      style="overflow-x: scroll; font-size: .8em"
      :items="formattedRecords"
      :fields="fields"
      :sorter="true"
      hover
      :responsive="false"
      :items-per-page="5"
      :pagination="paginationEnabled"
      :items-per-page-select="true"
      :loading="records == null"
      :small="true"
      :cleaner="true"
      :active-page="activePage"
      @row-clicked="(item) => $emit('select-record', item)"
      @pages-change="(pages) => $emit('pages-change', pages)"
    >
      <template #contents="{ item }">
        <td class="py-2">
          <div class="d-flex flex-wrap" style="gap: 6px; max-width: 400px;">
            <CCard
              v-for="ct in item.contents"
              :key="ct.id"
              class="mb-0 shadow-sm border-light"
              style="min-width: 120px; flex: 0 1 auto;"
            >
              <div class="p-1 d-flex align-items-center">
                <div
                  class="bg-info font-weight-bold rounded px-1 mr-2 d-flex align-items-center justify-content-center"
                  style="font-size: 0.75em; min-width: 35px; height: 20px;"
                >
                  {{ ct.id }}
                </div>
                <div class="text-truncate small font-weight-bold" style="max-width: 100px;">
                  {{ ct.name }}
                </div>
              </div>
            </CCard>
            <span v-if="!item.contents || item.contents.length === 0" class="text-muted small">No Content</span>
          </div>
        </td>
      </template>

      <template #raw_data="{ item }">
        <td>
          <pre class="mb-0 bg-light p-2 rounded text-dark border" style="max-height: 150px; overflow-y: auto;">
            <code>{{ JSON.stringify(item.raw_data, null, 2) }}</code>
          </pre>
        </td>
      </template>
    </CDataTable>
  </CCard>
</template>

<script>
export default {
  name: 'AttestationRecordsTable',
  props: {
    records: {
      type: Array,
      default: null,
    },
    paginationEnabled: {
      type: Boolean,
      default: false,
    },
    activePage: {
      type: Number,
      default: 1,
    },
  },
  data() {
    return {
      fields: [
        { key: 'unique_id', label: 'Device Unique ID', _style: 'width: 35%', sorter: false },
        { key: 'timestamp', label: 'Last Seen', _style: 'width: 20%' },
        { key: 'exp_datetime', label: 'Expired At', _style: 'width: 20%' },
        { key: 'raw_data', label: 'Raw Data', sorter: false },
      ],
    };
  },
  computed: {
    formattedRecords() {
      if (!this.records) return [];

      const now = Date.now();
      return this.records
        .map((row) => {
          const expiresAtMs = (row.claims?.request_details?.exp || 0) * 1000;
          const isExpired = expiresAtMs > 0 && expiresAtMs < now;

          return {
            ...row,
            unique_id: row.claims?.device_state?.unique_id || 'N/A',
            display_time: new Date(row.timestamp).toLocaleString(),
            exp: row.claims?.request_details?.exp,
            exp_datetime: new Date((row.claims?.request_details?.exp || 0) * 1000).toLocaleString(),
            raw_data: row,
            _classes: isExpired ? 'text-muted opacity-50 bg-light' : '',
          };
        })
        .sort((a, b) => b.exp - a.exp);
    },
  },
};
</script>
