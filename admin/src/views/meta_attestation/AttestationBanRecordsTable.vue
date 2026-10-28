<template>
  <CCard class="mb-4">
    <CCardHeader>
      <h3>Local Attestation Ban Records</h3>
      <h6>(Click to select Ban ID)</h6>
    </CCardHeader>
    <CDataTable
      style="overflow-x: scroll; font-size: .8em"
      :items="formattedBanRecords"
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
      @row-clicked="(item) => $emit('select-ban', item)"
      @pages-change="(pages) => $emit('pages-change', pages)"
    >
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
  name: 'AttestationBanRecordsTable',
  props: {
    records: {
      type: Array,
      default: null,
    },
    paginationEnabled: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      fields: [
        { key: 'action', label: 'Action', _style: 'width: 15%', sorter: false },
        { key: 'ban_id', label: 'Ban ID', _style: 'width: 25%', sorter: false },
        { key: 'ban_reason', label: 'Ban Reason', _style: 'width: 20%' },
        { key: 'remaining_time_in_minute', label: 'Remaining Time (min)', _style: 'width: 20%' },
        { key: 'raw_data', label: 'Raw Data', sorter: false },
      ],
    };
  },
  computed: {
    formattedBanRecords() {
      if (!this.records) return [];

      return this.records
        .map((row) => ({
          ...row,
          raw_data: row,
        }))
        .sort((a, b) => (b.exp || 0) - (a.exp || 0));
    },
  },
};
</script>
