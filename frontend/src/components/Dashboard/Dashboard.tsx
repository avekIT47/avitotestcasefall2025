import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
  UsersIcon,
  CodeBracketIcon,
  CheckCircleIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';
import {
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { apiService } from '../../services/api';
import { formatRelativeTime } from '../../utils';
import { useAppStore } from '../../store';
import { motion } from 'framer-motion';

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const { language } = useAppStore();

  const { data: statistics, isLoading: statsLoading } = useQuery({
    queryKey: ['statistics'],
    queryFn: () => apiService.getStatistics(),
  });

  const { data: recentPRs } = useQuery({
    queryKey: ['recent-prs'],
    queryFn: () => apiService.getPullRequests({ status: 'open' }).then(data => data.slice(0, 5)),
  });

  const statCards = [
    {
      title: t('dashboard.totalPRs'),
      value: statistics?.totalPRs || 0,
      icon: CodeBracketIcon,
      color: 'bg-blue-500',
    },
    {
      title: t('dashboard.openPRs'),
      value: statistics?.openPRs || 0,
      icon: ClockIcon,
      color: 'bg-yellow-500',
    },
    {
      title: t('dashboard.mergedPRs'),
      value: statistics?.mergedPRs || 0,
      icon: CheckCircleIcon,
      color: 'bg-green-500',
    },
  ];

  const pieChartData = [
    { name: t('dashboard.openPRs'), value: statistics?.openPRs || 0 },
    { name: t('dashboard.mergedPRs'), value: statistics?.mergedPRs || 0 },
  ];

  const COLORS = ['#3B82F6', '#10B981'];

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
      },
    },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.3 },
    },
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          {t('dashboard.title')}
        </h1>
        <p className="mt-1 text-gray-600 dark:text-gray-400">
          {t('dashboard.welcome')}
        </p>
      </div>

      {/* Stats Grid */}
      <motion.div
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        {statCards.map((stat) => (
          <motion.div
            key={stat.title}
            variants={itemVariants}
            className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">{stat.title}</p>
                <p className="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
                  {statsLoading ? (
                    <span className="inline-block w-16 h-7 bg-gray-200 dark:bg-gray-700 animate-pulse rounded" />
                  ) : (
                    stat.value
                  )}
                </p>
              </div>
              <div className={`p-3 rounded-full ${stat.color} bg-opacity-10`}>
                <stat.icon className={`h-6 w-6 ${stat.color.replace('bg-', 'text-')}`} />
              </div>
            </div>
          </motion.div>
        ))}
      </motion.div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Team Statistics Bar Chart */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            {t('dashboard.teamStats')}
          </h2>
          {statistics?.teamStats && statistics.teamStats.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={statistics.teamStats}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis 
                  dataKey="team_name" 
                  stroke="#9CA3AF"
                  tick={{ fill: '#9CA3AF' }}
                />
                <YAxis 
                  stroke="#9CA3AF"
                  tick={{ fill: '#9CA3AF' }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1F2937',
                    border: '1px solid #374151',
                    borderRadius: '0.5rem',
                  }}
                  labelStyle={{ color: '#F9FAFB' }}
                />
                <Legend wrapperStyle={{ color: '#9CA3AF' }} />
                <Bar dataKey="prCount" fill="#3B82F6" name={t('pr.title')} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[300px] flex items-center justify-center text-gray-500 dark:text-gray-400">
              {t('common.noData')}
            </div>
          )}
        </motion.div>

        {/* PR Status Pie Chart */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.4 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            PR Status Distribution
          </h2>
          {pieChartData.some(d => d.value > 0) ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieChartData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={(entry) => `${entry.name}: ${entry.value}`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieChartData.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1F2937',
                    border: '1px solid #374151',
                    borderRadius: '0.5rem',
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[300px] flex items-center justify-center text-gray-500 dark:text-gray-400">
              {t('common.noData')}
            </div>
          )}
        </motion.div>
      </div>

      {/* Recent Activity and Most Active Reviewer */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent PRs */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            {t('dashboard.recentActivity')}
          </h2>
          <div className="space-y-3">
            {recentPRs && recentPRs.length > 0 ? (
              recentPRs.map((pr) => (
                <div
                  key={pr.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-gray-50 dark:bg-gray-900 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {pr.title}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {pr.author?.name || 'Unknown'} • {formatRelativeTime(pr.createdAt, language)}
                    </p>
                  </div>
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
                    {pr.status}
                  </span>
                </div>
              ))
            ) : (
              <p className="text-center text-gray-500 dark:text-gray-400">
                {t('pr.noPRs')}
              </p>
            )}
          </div>
        </motion.div>

        {/* Most Active Reviewer */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            {t('dashboard.mostActiveReviewer')}
          </h2>
          {statistics?.userStats && statistics.userStats.length > 0 ? (
            <div className="text-center py-8">
              <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-primary-100 dark:bg-primary-900/30 mb-4">
                <UsersIcon className="h-10 w-10 text-primary-600 dark:text-primary-400" />
              </div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                {statistics.userStats[0].userName}
              </h3>
              <p className="mt-2 text-3xl font-bold text-primary-600 dark:text-primary-400">
                {statistics.userStats[0].assignmentCount}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">Reviews assigned</p>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              {t('common.noData')}
            </div>
          )}
        </motion.div>
      </div>
    </div>
  );
};

export default Dashboard;
