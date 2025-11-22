import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { apiService } from '../../services/api';
import { motion } from 'framer-motion';

const Statistics: React.FC = () => {
  const { t } = useTranslation();

  const { data: statistics, isLoading } = useQuery({
    queryKey: ['statistics'],
    queryFn: () => apiService.getStatistics(),
  });

  const { data: pullRequests } = useQuery({
    queryKey: ['pull-requests-stats'],
    queryFn: () => apiService.getPullRequests(),
  });

  // Prepare data for charts
  const monthlyData = React.useMemo(() => {
    if (!pullRequests) return [];
    
    const monthGroups: { [key: string]: { open: number; merged: number; created: number } } = {};
    
    pullRequests.forEach(pr => {
      const month = new Date(pr.createdAt).toLocaleDateString('en', { month: 'short', year: '2-digit' });
      if (!monthGroups[month]) {
        monthGroups[month] = { open: 0, merged: 0, created: 0 };
      }
      monthGroups[month].created++;
      if (pr.status === 'open') monthGroups[month].open++;
      if (pr.status === 'merged') monthGroups[month].merged++;
    });

    return Object.entries(monthGroups).map(([month, data]) => ({
      month,
      ...data,
    })).slice(-6); // Last 6 months
  }, [pullRequests]);

  const teamPerformance = React.useMemo(() => {
    if (!statistics || !statistics.teamStats) return [];

    return statistics.teamStats.map(team => ({
      team: team.teamName,
      prs: team.prCount,
    }));
  }, [statistics]);

  const statusDistribution = [
    { name: 'Open', value: statistics?.openPRs || 0, color: '#3B82F6' },
    { name: 'Merged', value: statistics?.mergedPRs || 0, color: '#10B981' },
    { 
      name: 'Closed', 
      value: (statistics?.totalPRs || 0) - (statistics?.openPRs || 0) - (statistics?.mergedPRs || 0),
      color: '#EF4444'
    },
  ];

  const radarData = React.useMemo(() => {
    if (!statistics || !statistics.teamStats) return [];

    return statistics.teamStats.slice(0, 5).map(team => ({
      team: team.teamName,
      A: team.prCount,
      fullMark: 150,
    }));
  }, [statistics]);

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

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          {t('nav.statistics')}
        </h1>
        <p className="mt-1 text-gray-600 dark:text-gray-400">
          Detailed analytics and insights
        </p>
      </div>

      {/* Key Metrics */}
      <motion.div
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        {[
          {
            title: 'Merge Rate',
            value: statistics?.totalPRs
              ? `${((statistics.mergedPRs / statistics.totalPRs) * 100).toFixed(1)}%`
              : '0%',
            subtitle: statistics?.mergedPRs 
              ? `${statistics.mergedPRs} of ${statistics.totalPRs} PRs merged`
              : 'No data',
          },
          {
            title: 'Open PRs',
            value: statistics?.openPRs || 0,
            subtitle: statistics?.totalPRs
              ? `${((statistics.openPRs / statistics.totalPRs) * 100).toFixed(0)}% of total`
              : 'No data',
          },
          {
            title: 'Total Teams',
            value: statistics?.teamStats?.length || 0,
            subtitle: statistics?.teamStats?.length ? 'Active teams' : 'No teams yet',
          },
          {
            title: 'Active Reviewers',
            value: statistics?.userStats?.length || 0,
            subtitle: statistics?.userStats?.length ? 'With assignments' : 'No reviewers yet',
          },
        ].map((metric, index) => (
          <motion.div
            key={index}
            variants={itemVariants}
            className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
          >
            <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{metric.title}</p>
            <div className="mt-2">
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{metric.value}</p>
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-500">
                {metric.subtitle}
              </p>
            </div>
          </motion.div>
        ))}
      </motion.div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Monthly Trend */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Monthly PR Trends
          </h2>
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={monthlyData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis dataKey="month" stroke="#9CA3AF" />
              <YAxis stroke="#9CA3AF" />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#1F2937',
                  border: '1px solid #374151',
                  borderRadius: '0.5rem',
                }}
              />
              <Legend />
              <Area type="monotone" dataKey="created" stackId="1" stroke="#8B5CF6" fill="#8B5CF6" fillOpacity={0.6} />
              <Area type="monotone" dataKey="merged" stackId="1" stroke="#10B981" fill="#10B981" fillOpacity={0.6} />
              <Area type="monotone" dataKey="open" stackId="1" stroke="#3B82F6" fill="#3B82F6" fillOpacity={0.6} />
            </AreaChart>
          </ResponsiveContainer>
        </motion.div>

        {/* Team Performance */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Team Performance
          </h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={teamPerformance} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis type="number" stroke="#9CA3AF" />
              <YAxis dataKey="team" type="category" stroke="#9CA3AF" width={80} />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#1F2937',
                  border: '1px solid #374151',
                  borderRadius: '0.5rem',
                }}
              />
              <Legend />
              <Bar dataKey="prs" fill="#3B82F6" name="Pull Requests" />
              <Bar dataKey="users" fill="#10B981" name="Active Users" />
            </BarChart>
          </ResponsiveContainer>
        </motion.div>

        {/* Status Distribution */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            PR Status Distribution
          </h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={statusDistribution}
                cx="50%"
                cy="50%"
                outerRadius={100}
                fill="#8884d8"
                dataKey="value"
                label={({ name, percent }) => `${name} ${((percent || 0) * 100).toFixed(0)}%`}
              >
                {statusDistribution.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
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
        </motion.div>

        {/* Radar Chart */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
        >
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Team Activity Radar
          </h2>
          <ResponsiveContainer width="100%" height={300}>
            <RadarChart data={radarData}>
              <PolarGrid stroke="#374151" />
              <PolarAngleAxis dataKey="team" stroke="#9CA3AF" />
              <PolarRadiusAxis angle={90} domain={[0, 150]} stroke="#9CA3AF" />
              <Radar name="PRs" dataKey="A" stroke="#3B82F6" fill="#3B82F6" fillOpacity={0.6} />
              <Radar name="Users x10" dataKey="B" stroke="#10B981" fill="#10B981" fillOpacity={0.6} />
              <Legend />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#1F2937',
                  border: '1px solid #374151',
                  borderRadius: '0.5rem',
                }}
              />
            </RadarChart>
          </ResponsiveContainer>
        </motion.div>
      </div>

      {/* Additional Stats */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
        className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
      >
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Top Contributors
        </h2>
        <div className="space-y-4">
          {statistics?.userStats && statistics.userStats.length > 0 && (
            <div className="flex items-center justify-between p-4 bg-gradient-to-r from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 rounded-lg">
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Most Active Reviewer</p>
                <p className="text-lg font-semibold text-gray-900 dark:text-white">
                  {statistics.userStats[0].userName}
                </p>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold text-primary-600 dark:text-primary-400">
                  {statistics.userStats[0].assignmentCount}
                </p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Assignments</p>
              </div>
            </div>
          )}
        </div>
      </motion.div>
    </div>
  );
};

export default Statistics;
