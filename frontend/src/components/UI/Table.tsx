import React, { useState, useMemo } from 'react';
import { ChevronUpIcon, ChevronDownIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { cn, debounce } from '../../utils';
import { useTranslation } from 'react-i18next';
import Button from './Button';

interface Column<T> {
  key: string;
  header: string;
  accessor: (item: T) => React.ReactNode;
  sortable?: boolean;
  width?: string;
  // Добавляем опциональные функции для получения реального значения для поиска и сортировки
  searchValue?: (item: T) => string | number;
  sortValue?: (item: T) => string | number | Date | null;
}

interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  loading?: boolean;
  searchable?: boolean;
  searchPlaceholder?: string;
  selectable?: boolean;
  selectedRows?: Set<number>;
  onSelectionChange?: (selected: Set<number>) => void;
  isRowSelectable?: (item: T, index: number) => boolean;
  onRowClick?: (item: T, index: number) => void;
  emptyMessage?: string;
  pagination?: boolean;
  pageSize?: number;
  className?: string;
  bulkActions?: React.ReactNode;
}

function Table<T extends { id?: number }>({
  columns,
  data,
  loading = false,
  searchable = false,
  searchPlaceholder,
  selectable = false,
  selectedRows = new Set(),
  onSelectionChange,
  isRowSelectable,
  onRowClick,
  emptyMessage,
  pagination = false,
  pageSize = 10,
  className,
  bulkActions,
}: TableProps<T>) {
  const { t } = useTranslation();
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: 'asc' | 'desc' } | null>(null);
  const [currentPage, setCurrentPage] = useState(1);

  // Filter data based on search
  const filteredData = useMemo(() => {
    if (!searchTerm) return data;
    
    return data.filter(item =>
      columns.some(column => {
        // Используем searchValue если указано, иначе пытаемся получить значение из самого объекта по ключу
        let value: any;
        if (column.searchValue) {
          value = column.searchValue(item);
        } else if (item && typeof item === 'object' && column.key in item) {
          value = (item as any)[column.key];
        } else {
          value = column.accessor(item);
        }
        
        if (value === null || value === undefined) return false;
        return value.toString().toLowerCase().includes(searchTerm.toLowerCase());
      })
    );
  }, [data, searchTerm, columns]);

  // Sort data
  const sortedData = useMemo(() => {
    if (!sortConfig) return filteredData;

    const sorted = [...filteredData].sort((a, b) => {
      const column = columns.find(col => col.key === sortConfig.key);
      if (!column) return 0;

      // Используем sortValue если указано, иначе пытаемся получить значение из самого объекта по ключу
      let aValue: any;
      let bValue: any;
      
      if (column.sortValue) {
        aValue = column.sortValue(a);
        bValue = column.sortValue(b);
      } else if (a && typeof a === 'object' && column.key in a) {
        aValue = (a as any)[column.key];
        bValue = (b as any)[column.key];
      } else {
        aValue = column.accessor(a);
        bValue = column.accessor(b);
      }

      if (aValue === null || aValue === undefined) return 1;
      if (bValue === null || bValue === undefined) return -1;

      // Обработка дат
      if (aValue instanceof Date && bValue instanceof Date) {
        return sortConfig.direction === 'asc' 
          ? aValue.getTime() - bValue.getTime()
          : bValue.getTime() - aValue.getTime();
      }

      // Обработка строк (игнорируем регистр)
      if (typeof aValue === 'string' && typeof bValue === 'string') {
        const comparison = aValue.toLowerCase().localeCompare(bValue.toLowerCase());
        return sortConfig.direction === 'asc' ? comparison : -comparison;
      }

      // Обработка чисел и других типов
      if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1;
      if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1;
      return 0;
    });

    return sorted;
  }, [filteredData, sortConfig, columns]);

  // Paginate data
  const paginatedData = useMemo(() => {
    if (!pagination) return sortedData;

    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    return sortedData.slice(startIndex, endIndex);
  }, [sortedData, pagination, currentPage, pageSize]);

  const totalPages = Math.ceil(sortedData.length / pageSize);

  const handleSort = (key: string) => {
    setSortConfig(current => {
      if (!current || current.key !== key) {
        return { key, direction: 'asc' };
      }
      if (current.direction === 'asc') {
        return { key, direction: 'desc' };
      }
      return null;
    });
  };

  const handleSelectAll = () => {
    if (selectedRows.size === paginatedData.length) {
      onSelectionChange?.(new Set());
    } else {
      const allIds = new Set(paginatedData.map((_, index) => index));
      onSelectionChange?.(allIds);
    }
  };

  const handleSelectRow = (index: number) => {
    const newSelected = new Set(selectedRows);
    if (newSelected.has(index)) {
      newSelected.delete(index);
    } else {
      newSelected.add(index);
    }
    onSelectionChange?.(newSelected);
  };

  const debouncedSearch = useMemo(
    () => debounce((value: string) => setSearchTerm(value), 300),
    []
  );

  return (
    <div className={cn('space-y-4', className)}>
      {/* Search and bulk actions */}
      <div className="flex flex-col sm:flex-row justify-between gap-4">
        {searchable && (
          <div className="relative flex-1 max-w-md">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              placeholder={searchPlaceholder || t('common.search')}
              onChange={(e) => debouncedSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
        )}
        {bulkActions && selectedRows.size > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {t('common.itemsSelected', { count: selectedRows.size })}
            </span>
            {bulkActions}
          </div>
        )}
      </div>

      {/* Table */}
      <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            <tr>
              {selectable && (
                <th className="px-6 py-3 w-12">
                  <input
                    type="checkbox"
                    checked={selectedRows.size === paginatedData.length && paginatedData.length > 0}
                    onChange={handleSelectAll}
                    className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
                  />
                </th>
              )}
              {columns.map(column => (
                <th
                  key={column.key}
                  className={cn(
                    'px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider',
                    column.sortable && 'cursor-pointer hover:text-gray-700 dark:hover:text-gray-200',
                    column.width
                  )}
                  onClick={() => column.sortable && handleSort(column.key)}
                >
                  <div className="flex items-center gap-1">
                    {column.header}
                    {column.sortable && sortConfig?.key === column.key && (
                      <>
                        {sortConfig.direction === 'asc' ? (
                          <ChevronUpIcon className="h-4 w-4" />
                        ) : (
                          <ChevronDownIcon className="h-4 w-4" />
                        )}
                      </>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
            {loading ? (
              <tr>
                <td colSpan={columns.length + (selectable ? 1 : 0)} className="px-6 py-8 text-center">
                  <div className="flex justify-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                  </div>
                </td>
              </tr>
            ) : paginatedData.length === 0 ? (
              <tr>
                <td colSpan={columns.length + (selectable ? 1 : 0)} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                  {emptyMessage || t('common.noData')}
                </td>
              </tr>
            ) : (
              paginatedData.map((item, index) => {
                const rowSelectable = isRowSelectable ? isRowSelectable(item, index) : true;
                return (
                  <tr
                    key={item.id || index}
                    className={cn(
                      'hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors',
                      onRowClick && 'cursor-pointer',
                      selectedRows.has(index) && 'bg-primary-50 dark:bg-primary-900/20',
                      !rowSelectable && 'opacity-60'
                    )}
                    onClick={() => onRowClick?.(item, index)}
                  >
                    {selectable && (
                      <td className="px-6 py-4 w-12">
                        <input
                          type="checkbox"
                          checked={selectedRows.has(index)}
                          disabled={!rowSelectable}
                          onChange={(e) => {
                            e.stopPropagation();
                            handleSelectRow(index);
                          }}
                          onClick={(e) => e.stopPropagation()}
                          className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                        />
                      </td>
                    )}
                    {columns.map(column => (
                      <td key={column.key} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                        {column.accessor(item)}
                      </td>
                    ))}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {pagination && totalPages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-gray-700 dark:text-gray-300">
            {t('common.showing')} {(currentPage - 1) * pageSize + 1} - {Math.min(currentPage * pageSize, sortedData.length)} {t('common.of')} {sortedData.length}
          </span>
          <div className="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              disabled={currentPage === 1}
              onClick={() => setCurrentPage(prev => prev - 1)}
            >
              {t('common.previous')}
            </Button>
            {Array.from({ length: totalPages }, (_, i) => i + 1)
              .filter(page => page === 1 || page === totalPages || Math.abs(page - currentPage) <= 1)
              .map((page, index, array) => (
                <React.Fragment key={page}>
                  {index > 0 && array[index - 1] !== page - 1 && <span className="px-2">...</span>}
                  <Button
                    variant={page === currentPage ? 'primary' : 'ghost'}
                    size="sm"
                    onClick={() => setCurrentPage(page)}
                  >
                    {page}
                  </Button>
                </React.Fragment>
              ))}
            <Button
              variant="secondary"
              size="sm"
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage(prev => prev + 1)}
            >
              {t('common.next')}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

export default Table;
