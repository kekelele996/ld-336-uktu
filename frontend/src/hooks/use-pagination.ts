import { signal } from '@angular/core';

// 分页工具（Angular signal 版），供所有列表页复用。
export function usePagination(defaultSize = 10) {
  const page = signal(1);
  const pageSize = signal(defaultSize);
  const total = signal(0);
  const loading = signal(false);

  function setPage(p: number): void {
    page.set(p);
  }

  function setTotal(t: number): void {
    total.set(t);
  }

  function setLoading(l: boolean): void {
    loading.set(l);
  }

  function reset(): void {
    page.set(1);
    total.set(0);
  }

  return { page, pageSize, total, loading, setPage, setTotal, setLoading, reset };
}
