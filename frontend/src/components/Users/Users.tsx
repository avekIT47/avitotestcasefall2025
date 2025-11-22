import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { PlusIcon, PencilIcon, UserIcon, CheckIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { toast } from 'react-hot-toast';
import { apiService } from '../../services/api';
import type { User, CreateUserRequest, UpdateUserRequest } from '../../types';
import Table from '../UI/Table';
import Button from '../UI/Button';
import Modal from '../UI/Modal';
import { formatDate, exportToCSV, getStatusColor } from '../../utils';
import { useAppStore } from '../../store';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';


const Users: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { language, selectedRows, setSelectedRows, clearSelectedRows } = useAppStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all');
  const [formData, setFormData] = useState<CreateUserRequest>({
    username: '',
    name: '',
    isActive: true,
  });

  // Fetch users with filter
  const { data: users, isLoading } = useQuery({
    queryKey: ['users', filterStatus],
    queryFn: () => {
      if (filterStatus === 'all') {
        return apiService.getUsers();
      }
      return apiService.getUsers({ isActive: filterStatus === 'active' });
    },
  });

  // Ensure users is never null/undefined
  const usersData = users || [];

  // Create user mutation
  const createUserMutation = useMutation({
    mutationFn: (data: CreateUserRequest) => apiService.createUser(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success(t('messages.userCreated'));
      setIsCreateModalOpen(false);
      setFormData({ username: '', name: '', isActive: true });
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Update user mutation
  const updateUserMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateUserRequest }) =>
      apiService.updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success(t('messages.userUpdated'));
      setIsEditModalOpen(false);
      setSelectedUser(null);
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Bulk deactivate
  const handleBulkDeactivate = async () => {
    const selectedUsers = Array.from(selectedRows).map(index => usersData[index]).filter(Boolean);
    
    for (const user of selectedUsers) {
      if (user.isActive) {
        await apiService.updateUser(user.id, { isActive: false });
      }
    }
    
    queryClient.invalidateQueries({ queryKey: ['users'] });
    toast.success(t('messages.usersDeactivated'));
    clearSelectedRows();
  };

  // Bulk activate
  const handleBulkActivate = async () => {
    const selectedUsers = Array.from(selectedRows).map(index => usersData[index]).filter(Boolean);
    
    for (const user of selectedUsers) {
      if (!user.isActive) {
        await apiService.updateUser(user.id, { isActive: true });
      }
    }
    
    queryClient.invalidateQueries({ queryKey: ['users'] });
    toast.success('Users activated successfully');
    clearSelectedRows();
  };

  // Export to CSV
  const handleExportCSV = () => {
    const data = usersData.map(user => ({
      id: user.id,
      username: user.username,
      name: user.name,
      status: user.isActive ? 'Active' : 'Inactive',
      teams: user.teams?.filter(t => t).map(t => t.name).join(', ') || '',
      createdAt: formatDate(user.createdAt, language),
      updatedAt: formatDate(user.updatedAt, language),
    }));
    exportToCSV(data, ['id', 'username', 'name', 'status', 'teams', 'createdAt', 'updatedAt'], 'users.csv');
    toast.success(t('messages.exportSuccess'));
  };

  // Export to PDF
  const handleExportPDF = () => {
    const doc = new jsPDF();
    
    // Add title
    doc.setFontSize(18);
    doc.text('Users Report', 14, 15);
    doc.setFontSize(11);
    doc.text(new Date().toLocaleString(), 14, 25);
    
    // Prepare table data
    const tableData = usersData.map(user => [
      user.id.toString(),
      user.username,
      user.name,
      user.isActive ? 'Active' : 'Inactive',
      user.teams?.filter(t => t).map(t => t.name).join(', ') || '-',
    ]);
    
    // Add table
    autoTable(doc, {
      head: [['ID', 'Username', 'Name', 'Status', 'Teams']],
      body: tableData,
      startY: 35,
      theme: 'grid',
      styles: { fontSize: 10 },
      headStyles: { fillColor: [59, 130, 246] },
    });
    
    // Save PDF
    doc.save('users.pdf');
    toast.success(t('messages.exportSuccess'));
  };

  const columns = [
    {
      key: 'id',
      header: 'ID',
      accessor: (user: User) => user.id,
      searchValue: (user: User) => `${user.id} ${user.username} ${user.name}`,
      sortValue: (user: User) => user.id,
      sortable: true,
      width: 'w-16',
    },
    {
      key: 'username',
      header: t('users.username'),
      accessor: (user: User) => (
        <div className="flex items-center gap-2">
          <UserIcon className="h-5 w-5 text-gray-400" />
          <span className="font-medium">{user.username}</span>
        </div>
      ),
      searchValue: (user: User) => `${user.id} ${user.username} ${user.name}`,
      sortValue: (user: User) => user.username,
      sortable: true,
    },
    {
      key: 'name',
      header: t('users.name'),
      accessor: (user: User) => user.name,
      searchValue: (user: User) => `${user.id} ${user.username} ${user.name}`,
      sortValue: (user: User) => user.name,
      sortable: true,
    },
    {
      key: 'status',
      header: t('users.status'),
      accessor: (user: User) => (
        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(user.isActive ? 'active' : 'inactive')}`}>
          {user.isActive ? (
            <CheckIcon className="h-3 w-3" />
          ) : (
            <XMarkIcon className="h-3 w-3" />
          )}
          {user.isActive ? t('users.active') : t('users.inactive')}
        </span>
      ),
      searchValue: (user: User) => user.isActive ? 'active' : 'inactive',
      sortValue: (user: User) => user.isActive ? 1 : 0,
      sortable: true,
    },
    {
      key: 'teams',
      header: t('users.teams'),
      accessor: (user: User) => (
        <div className="flex flex-wrap gap-1">
          {user.teams && user.teams.length > 0 ? (
            user.teams.map(team => (
              <span
                key={team.id}
                className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
              >
                {team.name}
              </span>
            ))
          ) : (
            <span className="text-gray-400 text-sm">-</span>
          )}
        </div>
      ),
      searchValue: (user: User) => user.teams?.map(t => t.name).join(' ') || '',
      sortable: false,
    },
    {
      key: 'actions',
      header: t('users.actions'),
      accessor: (user: User) => (
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="ghost"
            icon={<PencilIcon className="h-4 w-4" />}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedUser(user);
              setFormData({
                username: user.username,
                name: user.name,
                isActive: user.isActive,
              });
              setIsEditModalOpen(true);
            }}
          >
            {t('users.edit')}
          </Button>
          <Button
            size="sm"
            variant={user.isActive ? 'danger' : 'primary'}
            onClick={(e) => {
              e.stopPropagation();
              updateUserMutation.mutate({
                id: user.id,
                data: { isActive: !user.isActive },
              });
            }}
          >
            {user.isActive ? t('users.deactivate') : t('users.activate')}
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            {t('users.title')}
          </h1>
          <p className="mt-1 text-gray-600 dark:text-gray-400">
            Manage users and their permissions
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            onClick={handleExportCSV}
          >
            Export CSV
          </Button>
          <Button
            variant="secondary"
            onClick={handleExportPDF}
          >
            Export PDF
          </Button>
          <Button
            variant="primary"
            icon={<PlusIcon className="h-5 w-5" />}
            onClick={() => setIsCreateModalOpen(true)}
          >
            {t('users.createUser')}
          </Button>
        </div>
      </div>

      {/* Filter tabs */}
      <div className="flex gap-2 border-b border-gray-200 dark:border-gray-700">
        {(['all', 'active', 'inactive'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilterStatus(status)}
            className={`px-4 py-2 font-medium transition-colors border-b-2 ${
              filterStatus === status
                ? 'border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                : 'border-transparent text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
            }`}
          >
            {status === 'all' && t('users.allUsers')}
            {status === 'active' && t('users.activeOnly')}
            {status === 'inactive' && t('users.inactiveOnly')}
          </button>
        ))}
      </div>

      {/* Users Table */}
      <Table
        columns={columns}
        data={usersData}
        loading={isLoading}
        searchable
        searchPlaceholder="Search users..."
        selectable
        selectedRows={selectedRows}
        onSelectionChange={setSelectedRows}
        emptyMessage={t('users.noUsers')}
        pagination
        pageSize={10}
        bulkActions={
          <>
            {filterStatus !== 'inactive' && (
              <Button
                size="sm"
                variant="danger"
                onClick={handleBulkDeactivate}
              >
                {t('users.deactivate')}
              </Button>
            )}
            {filterStatus !== 'active' && (
              <Button
                size="sm"
                variant="primary"
                onClick={handleBulkActivate}
              >
                {t('users.activate')}
              </Button>
            )}
          </>
        }
      />

      {/* Create/Edit User Modal */}
      <Modal
        isOpen={isCreateModalOpen || isEditModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          setIsEditModalOpen(false);
          setSelectedUser(null);
          setFormData({ username: '', name: '', isActive: true });
        }}
        title={isEditModalOpen ? `Edit User - ${selectedUser?.name}` : t('users.createUser')}
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('users.username')}
            </label>
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              disabled={isEditModalOpen}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
              placeholder="Enter username"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('users.name')}
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Enter full name"
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="isActive"
              checked={formData.isActive}
              onChange={(e) => setFormData({ ...formData, isActive: e.target.checked })}
              className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
            />
            <label htmlFor="isActive" className="text-sm font-medium text-gray-700 dark:text-gray-300">
              {t('users.active')}
            </label>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsCreateModalOpen(false);
                setIsEditModalOpen(false);
                setSelectedUser(null);
                setFormData({ username: '', name: '', isActive: true });
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={() => {
                if (isEditModalOpen && selectedUser) {
                  updateUserMutation.mutate({
                    id: selectedUser.id,
                    data: { name: formData.name, isActive: formData.isActive },
                  });
                } else {
                  createUserMutation.mutate(formData);
                }
              }}
              loading={createUserMutation.isPending || updateUserMutation.isPending}
              disabled={!formData.username.trim() || !formData.name.trim()}
            >
              {t('common.save')}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Users;
