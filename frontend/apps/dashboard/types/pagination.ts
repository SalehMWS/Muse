export interface Paginated<TItem> {
  items: TItem[];
  cursor: string | null;
  hasNext: boolean;
  total?: number;
}

export interface CursorParams {
  cursor?: string | null;
  limit?: number;
}
