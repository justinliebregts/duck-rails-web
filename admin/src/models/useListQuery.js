import { reactive, watch } from '@vue/composition-api';

export function useListQuery(fetchFn, defaultLimit = 100) {
    const query = reactive({
        page_per_count: defaultLimit,
        page: 1,
        q: '',
        exclude_deleted: false,
        player_uuid: '',
        eq_id: [],
        item_id: [],
        sort: '',
        order: '',
    });

    let searchTimer = null;

    const clearTimer = () => {
        if (searchTimer) {
            clearTimeout(searchTimer);
            searchTimer = null;
        }
    };

    let isResetting = false;
    const reset = () => {
        isResetting = true;
        clearTimer();

        query.page = 1;
        query.q = '';
        query.exclude_deleted = false;
        query.player_uuid = '';
        query.eq_id = [];
        query.item_id = [];
        query.sort = '';
        query.order = '';

        setTimeout(() => {
            isResetting = false;
        }, 0);
    };

    const setPageSize = (newLimit) => {
        query.page_per_count = Number(newLimit);
        query.page = 1; // 自動歸零
    };

    const submitAndFetch = () => {
        clearTimer();

        if (query.page !== 1) {
            query.page = 1; // 修改 page 觸發 watch 執行 fetchFn
            return;
        }

        fetchFn(); // 頁碼未變動，手動執行
    };

    const handleSearchChange = (delay = 400) => {
        clearTimer();
        searchTimer = setTimeout(() => {
            submitAndFetch();
        }, delay);
    };

    const resetAndFetch = () => {
        reset();
        fetchFn();
    };

    const toParams = () => {
        const cleanParams = {};
        for (const [key, value] of Object.entries(query)) {
            if (value !== null && value !== undefined && value !== '') {
                if (Array.isArray(value)) {
                    if (value.length > 0) cleanParams[key] = value;
                } else {
                    cleanParams[key] = value;
                }
            }
        }
        return cleanParams;
    };


    // 監聽通用分頁與筆數變更 -> 自動觸發 API
    watch(
        () => [query.page, query.page_per_count],
        function ([newPage, newCount], [oldPage, oldCount]) {
            // console.log("[page]", query.page, " ", newPage, "->", oldPage);
            // console.log("[page_count]", query.page_per_count, "", newCount, "->", oldCount);
            if (isResetting)
                return;

            if (newPage !== oldPage || newCount !== oldCount) {
                fetchFn();
            }
        }
    );

    // 監聽搜尋關鍵字 q -> 套用防抖
    watch(
        () => query.q,
        (newVal, oldVal) => {
            // console.log("[q]", newVal, oldVal);
            if (newVal === oldVal)
                return;

            if (isResetting)
                return;

            handleSearchChange(400);
        }
    );

    return {
        query,
        searchTimer,

        clearTimer,
        reset,
        setPageSize,
        submitAndFetch,
        handleSearchChange,
        resetAndFetch,
        toParams
    };
}