export class ListQuery {
    constructor(defaultLimit = 100) {
        this.page_per_count = defaultLimit;

        this.page = 1;
        this.q = '';
        this.exclude_deleted = false;
        this.player_uuid = '';
        this.eq_id = [];
        this.item_id = [];
        this.sort = '';
        this.order = '';
        this.searchTimer = null;
    }

    clearTimer() {
        if (this.searchTimer) {
            clearTimeout(this.searchTimer);
            this.searchTimer = null;
        }
    }

    reset() {
        this.clearTimer();

        this.page = 1;
        this.q = '';
        this.exclude_deleted = false;
        this.player_uuid = '';
        this.eq_id = [];
        this.item_id = [];
        this.sort = '';
        this.order = '';
    }

    submitAndFetch(fetchFn) {
        this.clearTimer();

        if (this.page !== 1) {
            this.page = 1; // 觸發 Vue watch['banQuery.page'] -> 自動呼叫 fetchFn
            return;
        }

        fetchFn();     // 頁碼未變動，手動執行
    }

    handleSearchChange(fetchFn, delay = 400) {
        this.clearTimer();
        this.searchTimer = setTimeout(() => {
            this.submitAndFetch(fetchFn);
        }, delay);
    }

    resetAndFetch(fetchFn) {
        const wasFirstPage = (this.page === 1);

        this.reset();

        if (!wasFirstPage) {
            // 若原本不是第 1 頁，this.reset() 已將 page 改為 1，自動觸發 watcher 執行，無需額外動作
            return
        }

        fetchFn(); // 原本就是第 1 頁，page 未改變，手動執行
    }
}
