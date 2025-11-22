import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
  PlusIcon,
  ArrowPathIcon,
  CheckCircleIcon,
  CodeBracketIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';
import { toast } from 'react-hot-toast';
import { apiService } from '../../services/api';
import type {
  PullRequest,
  CreatePullRequestRequest,
  ReassignReviewerRequest,
} from '../../types';
import Table from '../UI/Table';
import Button from '../UI/Button';
import Modal from '../UI/Modal';
import { formatDate, formatRelativeTime, getStatusColor, exportToCSV } from '../../utils';
import { useAppStore } from '../../store';

const PullRequests: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { language, selectedRows, setSelectedRows, clearSelectedRows } = useAppStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isAddReviewerModalOpen, setIsAddReviewerModalOpen] = useState(false);
  const [isReassignModalOpen, setIsReassignModalOpen] = useState(false);
  const [selectedPR, setSelectedPR] = useState<PullRequest | null>(null);
  const [filterStatus, setFilterStatus] = useState<'all' | 'open' | 'merged' | 'closed'>('all');
  const [newReviewerId, setNewReviewerId] = useState<number | null>(null);
  const [formData, setFormData] = useState<CreatePullRequestRequest>({
    title: '',
    authorId: 0,
    teamId: 0,
  });
  const [reassignData, setReassignData] = useState<ReassignReviewerRequest>({
    oldReviewerId: 0,
    newReviewerId: undefined,
  });

  // Fetch pull requests with filter
  const { data: pullRequests, isLoading } = useQuery({
    queryKey: ['pull-requests', filterStatus],
    queryFn: () => {
      if (filterStatus === 'all') {
        return apiService.getPullRequests();
      }
      return apiService.getPullRequests({ status: filterStatus });
    },
  });

  // Fetch users and teams for dropdowns
  const { data: users } = useQuery({
    queryKey: ['users', 'active'],
    queryFn: () => apiService.getUsers({ isActive: true }),
  });

  const { data: teams } = useQuery({
    queryKey: ['teams'],
    queryFn: () => apiService.getTeams(),
  });

  // Ensure arrays are never null/undefined
  const pullRequestsData = pullRequests || [];
  const usersData = users || [];
  const teamsData = teams || [];

  // Create PR mutation
  const createPRMutation = useMutation({
    mutationFn: (data: CreatePullRequestRequest) => apiService.createPullRequest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
      queryClient.invalidateQueries({ queryKey: ['statistics'] });
      toast.success(t('messages.prCreated'));
      setIsCreateModalOpen(false);
      setFormData({ title: '', authorId: 0, teamId: 0 });
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Add reviewer mutation
  const addReviewerMutation = useMutation({
    mutationFn: ({ prId, reviewerId }: { prId: number; reviewerId: number }) =>
      apiService.addReviewer(prId, reviewerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
      toast.success('Reviewer added successfully');
      setIsAddReviewerModalOpen(false);
      setSelectedPR(null);
      setNewReviewerId(null);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to add reviewer');
    },
  });

  // Reassign reviewer mutation
  const reassignReviewerMutation = useMutation({
    mutationFn: ({ prId, data }: { prId: number; data: ReassignReviewerRequest }) =>
      apiService.reassignReviewer(prId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
      toast.success(t('messages.reviewerReassigned'));
      setIsReassignModalOpen(false);
      setSelectedPR(null);
      setReassignData({ oldReviewerId: 0, newReviewerId: undefined });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || t('messages.errorOccurred'));
    },
  });

  // Merge PR mutation
  const mergePRMutation = useMutation({
    mutationFn: (prId: number) => apiService.mergePullRequest(prId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
      queryClient.invalidateQueries({ queryKey: ['statistics'] });
      toast.success(t('messages.prMerged'));
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Close PR mutation
  const closePRMutation = useMutation({
    mutationFn: (prId: number) => apiService.closePullRequest(prId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
      queryClient.invalidateQueries({ queryKey: ['statistics'] });
      toast.success('Pull request closed successfully');
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Bulk merge
  const handleBulkMerge = async () => {
    const selectedPRs = Array.from(selectedRows)
      .map(index => pullRequestsData[index])
      .filter(pr => pr && pr.status === 'open');
    
    if (selectedPRs.length === 0) {
      toast.error('No open pull requests selected');
      return;
    }
    
    for (const pr of selectedPRs) {
      await apiService.mergePullRequest(pr.id);
    }
    
    queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
    queryClient.invalidateQueries({ queryKey: ['statistics'] });
    toast.success(`${selectedPRs.length} pull request(s) merged successfully`);
    clearSelectedRows();
  };

  // Bulk close
  const handleBulkClose = async () => {
    const selectedPRs = Array.from(selectedRows)
      .map(index => pullRequestsData[index])
      .filter(pr => pr && pr.status === 'open');
    
    if (selectedPRs.length === 0) {
      toast.error('No open pull requests selected');
      return;
    }
    
    for (const pr of selectedPRs) {
      await apiService.closePullRequest(pr.id);
    }
    
    queryClient.invalidateQueries({ queryKey: ['pull-requests'] });
    queryClient.invalidateQueries({ queryKey: ['statistics'] });
    toast.success(`${selectedPRs.length} pull request(s) closed successfully`);
    clearSelectedRows();
  };

  // Export to CSV
  const handleExport = () => {
    const data = pullRequestsData.map(pr => ({
      id: pr.id,
      title: pr.title,
      author: pr.author?.name || 'Unknown',
      team: pr.team?.name || 'Unknown',
      reviewers: pr.reviewers?.map(r => r?.name || 'Unknown').join(', ') || '',
      status: pr.status,
      createdAt: formatDate(pr.createdAt, language),
      mergedAt: pr.mergedAt ? formatDate(pr.mergedAt, language) : '-',
    }));
    exportToCSV(
      data,
      ['id', 'title', 'author', 'team', 'reviewers', 'status', 'createdAt', 'mergedAt'],
      'pull-requests.csv'
    );
    toast.success(t('messages.exportSuccess'));
  };

  // Count selected open PRs
  const selectedOpenPRsCount = Array.from(selectedRows)
    .map(index => pullRequestsData[index])
    .filter(pr => pr && pr.status === 'open')
    .length;

  const columns = [
    {
      key: 'id',
      header: 'ID',
      accessor: (pr: PullRequest) => pr.id,
      searchValue: (pr: PullRequest) => `${pr.id} ${pr.title} ${pr.author?.name || ''} ${pr.author?.username || ''} ${pr.team?.name || ''}`,
      sortValue: (pr: PullRequest) => pr.id,
      sortable: true,
      width: 'w-16',
    },
    {
      key: 'title',
      header: t('pr.prTitle'),
      accessor: (pr: PullRequest) => (
        <div className="flex items-center gap-2">
          <CodeBracketIcon className="h-5 w-5 text-gray-400" />
          <div>
            <p className="font-medium text-gray-900 dark:text-white">{pr.title}</p>
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {formatRelativeTime(pr.createdAt, language)}
            </p>
          </div>
        </div>
      ),
      searchValue: (pr: PullRequest) => `${pr.id} ${pr.title} ${pr.author?.name || ''} ${pr.author?.username || ''} ${pr.team?.name || ''}`,
      sortValue: (pr: PullRequest) => pr.title,
      sortable: true,
    },
    {
      key: 'author',
      header: t('pr.author'),
      accessor: (pr: PullRequest) => (
        <div>
          <p className="font-medium">{pr.author?.name || 'Unknown'}</p>
          <p className="text-xs text-gray-500 dark:text-gray-400">@{pr.author?.username || 'unknown'}</p>
        </div>
      ),
      searchValue: (pr: PullRequest) => `${pr.id} ${pr.title} ${pr.author?.name || ''} ${pr.author?.username || ''} ${pr.team?.name || ''}`,
      sortValue: (pr: PullRequest) => pr.author?.name || '',
      sortable: true,
    },
    {
      key: 'team',
      header: t('pr.team'),
      accessor: (pr: PullRequest) => (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400">
          {pr.team?.name || 'Unknown'}
        </span>
      ),
      searchValue: (pr: PullRequest) => `${pr.id} ${pr.title} ${pr.author?.name || ''} ${pr.author?.username || ''} ${pr.team?.name || ''}`,
      sortValue: (pr: PullRequest) => pr.team?.name || '',
      sortable: true,
    },
    {
      key: 'reviewers',
      header: t('pr.reviewers'),
      accessor: (pr: PullRequest) => (
        <div className="flex flex-wrap gap-1">
          {pr.reviewers && pr.reviewers.length > 0 ? (
            pr.reviewers.map(reviewer => (
              <span
                key={reviewer?.id}
                className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400"
              >
                {reviewer?.name || 'Unknown'}
              </span>
            ))
          ) : (
            <span className="text-gray-400 text-sm">No reviewers</span>
          )}
        </div>
      ),
      searchValue: (pr: PullRequest) => pr.reviewers?.map(r => r?.name || '').join(' ') || '',
      sortable: false,
    },
    {
      key: 'status',
      header: t('pr.status'),
      accessor: (pr: PullRequest) => (
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(pr.status)}`}>
          {pr.status === 'open' && t('pr.open')}
          {pr.status === 'merged' && t('pr.merged')}
          {pr.status === 'closed' && t('pr.closed')}
        </span>
      ),
      searchValue: (pr: PullRequest) => pr.status,
      sortValue: (pr: PullRequest) => pr.status,
      sortable: true,
    },
    {
      key: 'actions',
      header: t('pr.actions'),
      accessor: (pr: PullRequest) => (
        <div className="flex items-center gap-2">
          {pr.status === 'open' ? (
            <>
              <Button
                size="sm"
                variant="ghost"
                icon={<PlusIcon className="h-4 w-4" />}
                onClick={(e) => {
                  e.stopPropagation();
                  setSelectedPR(pr);
                  setIsAddReviewerModalOpen(true);
                }}
              >
                Add Reviewer
              </Button>
              {pr.reviewers && pr.reviewers.length > 0 && (
                <Button
                  size="sm"
                  variant="ghost"
                  icon={<ArrowPathIcon className="h-4 w-4" />}
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedPR(pr);
                    if (pr.reviewers && pr.reviewers[0]) {
                      setReassignData({
                        oldReviewerId: pr.reviewers[0].id,
                        newReviewerId: undefined,
                      });
                    }
                    setIsReassignModalOpen(true);
                  }}
                >
                  Replace
                </Button>
              )}
              <Button
                size="sm"
                variant="primary"
                icon={<CheckCircleIcon className="h-4 w-4" />}
                onClick={(e) => {
                  e.stopPropagation();
                  mergePRMutation.mutate(pr.id);
                }}
                loading={mergePRMutation.isPending}
              >
                {t('pr.merge')}
              </Button>
              <Button
                size="sm"
                variant="danger"
                icon={<XCircleIcon className="h-4 w-4" />}
                onClick={(e) => {
                  e.stopPropagation();
                  closePRMutation.mutate(pr.id);
                }}
                loading={closePRMutation.isPending}
              >
                Close
              </Button>
            </>
          ) : (
            <span className="text-sm font-medium">
              {pr.status === 'merged' && (
                <span className="text-green-600 dark:text-green-400">
                  ✓ Merged
                </span>
              )}
              {pr.status === 'closed' && (
                <span className="text-red-600 dark:text-red-400">
                  ✗ Closed
                </span>
              )}
            </span>
          )}
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
            {t('pr.title')}
          </h1>
          <p className="mt-1 text-gray-600 dark:text-gray-400">
            Manage pull requests and reviewers
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" onClick={handleExport}>
            {t('common.export')}
          </Button>
          <Button
            variant="primary"
            icon={<PlusIcon className="h-5 w-5" />}
            onClick={() => setIsCreateModalOpen(true)}
          >
            {t('pr.createPR')}
          </Button>
        </div>
      </div>

      {/* Filter tabs */}
      <div className="flex gap-2 border-b border-gray-200 dark:border-gray-700">
        {(['all', 'open', 'merged', 'closed'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilterStatus(status)}
            className={`px-4 py-2 font-medium transition-colors border-b-2 ${
              filterStatus === status
                ? 'border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                : 'border-transparent text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
            }`}
          >
            {status === 'all' && t('pr.allStatuses')}
            {status === 'open' && t('pr.open')}
            {status === 'merged' && t('pr.merged')}
            {status === 'closed' && t('pr.closed')}
          </button>
        ))}
      </div>

      {/* Pull Requests Table */}
      <Table
        columns={columns}
        data={pullRequestsData}
        loading={isLoading}
        searchable
        searchPlaceholder="Search pull requests..."
        selectable={filterStatus === 'all' || filterStatus === 'open'}
        selectedRows={selectedRows}
        onSelectionChange={setSelectedRows}
        isRowSelectable={(pr: PullRequest) => {
          // В режиме "все статусы" только открытые PR можно выбирать
          if (filterStatus === 'all') {
            return pr.status === 'open';
          }
          // В остальных вкладках все строки считаются "селектируемыми" 
          // (чтобы не применялась opacity), хотя чекбоксы там скрыты
          return true;
        }}
        emptyMessage={t('pr.noPRs')}
        pagination
        pageSize={10}
        bulkActions={
          (filterStatus === 'all' || filterStatus === 'open') && (
            <div className="flex gap-2">
              <Button 
                size="sm" 
                variant="primary" 
                onClick={handleBulkMerge}
                disabled={selectedOpenPRsCount === 0}
              >
                Merge Selected ({selectedOpenPRsCount})
              </Button>
              <Button 
                size="sm" 
                variant="danger" 
                onClick={handleBulkClose}
                icon={<XCircleIcon className="h-4 w-4" />}
                disabled={selectedOpenPRsCount === 0}
              >
                Close Selected ({selectedOpenPRsCount})
              </Button>
            </div>
          )
        }
      />

      {/* Create PR Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          setFormData({ title: '', authorId: 0, teamId: 0 });
        }}
        title={t('pr.createPR')}
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('pr.prTitle')}
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Enter PR title"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('pr.author')}
            </label>
            <select
              value={formData.authorId}
              onChange={(e) => setFormData({ ...formData, authorId: Number(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Select author...</option>
              {usersData.map((user) => (
                <option key={user.id} value={user.id}>
                  {user.name} (@{user.username})
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('pr.team')}
            </label>
            <select
              value={formData.teamId}
              onChange={(e) => setFormData({ ...formData, teamId: Number(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Select team...</option>
              {teamsData.map((team) => (
                <option key={team.id} value={team.id}>
                  {team.name}
                </option>
              ))}
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsCreateModalOpen(false);
                setFormData({ title: '', authorId: 0, teamId: 0 });
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={() => createPRMutation.mutate(formData)}
              loading={createPRMutation.isPending}
              disabled={!formData.title.trim() || !formData.authorId || !formData.teamId}
            >
              {t('common.save')}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Add Reviewer Modal */}
      <Modal
        isOpen={isAddReviewerModalOpen}
        onClose={() => {
          setIsAddReviewerModalOpen(false);
          setSelectedPR(null);
          setNewReviewerId(null);
        }}
        title={`Add Reviewer - ${selectedPR?.title}`}
        size="md"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Select a user to add as a reviewer for this pull request.
          </p>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Reviewer
            </label>
            <select
              value={newReviewerId || ''}
              onChange={(e) => setNewReviewerId(e.target.value ? Number(e.target.value) : null)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Select reviewer...</option>
              {usersData
                .filter(u => {
                  // Исключаем автора
                  if (u.id === selectedPR?.authorId) return false;
                  // Исключаем уже назначенных рецензентов
                  if (selectedPR?.reviewers?.some(r => r?.id === u.id)) return false;
                  return true;
                })
                .map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name} (@{user.username})
                  </option>
                ))}
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsAddReviewerModalOpen(false);
                setSelectedPR(null);
                setNewReviewerId(null);
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={() => {
                if (selectedPR && newReviewerId) {
                  addReviewerMutation.mutate({
                    prId: selectedPR.id,
                    reviewerId: newReviewerId,
                  });
                }
              }}
              loading={addReviewerMutation.isPending}
              disabled={!newReviewerId}
            >
              Add Reviewer
            </Button>
          </div>
        </div>
      </Modal>

      {/* Reassign Reviewer Modal */}
      <Modal
        isOpen={isReassignModalOpen}
        onClose={() => {
          setIsReassignModalOpen(false);
          setSelectedPR(null);
          setReassignData({ oldReviewerId: 0, newReviewerId: undefined });
        }}
        title={`Replace Reviewer - ${selectedPR?.title}`}
        size="md"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Replace an existing reviewer with a new one (or remove by leaving new reviewer empty).
          </p>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Reviewer to Replace
            </label>
            <select
              value={reassignData.oldReviewerId}
              onChange={(e) =>
                setReassignData({ ...reassignData, oldReviewerId: Number(e.target.value) })
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Select reviewer to replace...</option>
              {selectedPR?.reviewers?.map((reviewer) => (
                <option key={reviewer?.id} value={reviewer?.id}>
                  {reviewer?.name} (@{reviewer?.username})
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              New Reviewer (optional - leave empty to just remove)
            </label>
            <select
              value={reassignData.newReviewerId || ''}
              onChange={(e) =>
                setReassignData({
                  ...reassignData,
                  newReviewerId: e.target.value ? Number(e.target.value) : undefined,
                })
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Auto-assign from same team...</option>
              {usersData
                .filter(u => {
                  // Исключаем автора
                  if (u.id === selectedPR?.authorId) return false;
                  // Исключаем текущего заменяемого рецензента
                  if (u.id === reassignData.oldReviewerId) return false;
                  // Исключаем уже назначенных рецензентов
                  if (selectedPR?.reviewers?.some(r => r?.id === u.id)) return false;
                  return true;
                })
                .map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name} (@{user.username})
                  </option>
                ))}
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsReassignModalOpen(false);
                setSelectedPR(null);
                setReassignData({ oldReviewerId: 0, newReviewerId: undefined });
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={() => {
                if (selectedPR && reassignData.oldReviewerId) {
                  reassignReviewerMutation.mutate({
                    prId: selectedPR.id,
                    data: reassignData,
                  });
                }
              }}
              loading={reassignReviewerMutation.isPending}
              disabled={!reassignData.oldReviewerId}
            >
              Replace Reviewer
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default PullRequests;
