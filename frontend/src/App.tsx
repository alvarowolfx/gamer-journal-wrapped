import { AnimatePresence, motion } from 'framer-motion'
import {
  Calendar,
  Download,
  Eye,
  Gamepad2,
  Heart,
  Image as ImageIcon,
  Layers,
  Loader2,
  Monitor,
  Sparkles,
  Trophy
} from 'lucide-react'
import { useEffect, useState } from 'react'
import {
  Bar,
  LabelList,
  BarChart as ReBarChart,
  ResponsiveContainer,
  Tooltip,
  XAxis, YAxis,
  type RenderableText
} from 'recharts'
import { getChartUrl, getStats, type MostPlayedByNumGames, type MostPlayedByPlaytime, type MostPlayedGame, type StatsResponse } from './api'
import './index.css'

type DataType = MostPlayedByPlaytime | MostPlayedGame | MostPlayedByNumGames;

interface StatChartProps {
  title: string;
  icon: React.ReactNode;
  data: DataType[];
  dataKey: string;
  color: string;
  year: number;
  isMonths?: boolean;
  isStatus?: boolean;
  isLoading?: boolean;
}

const CustomTooltip = ({ active, payload, label, isStatus }: any) => {
  if (active && payload && payload.length) {
    return (
      <div className="custom-tooltip">
        <p className="label">{label || payload[0].payload.title}</p>
        <p className="value" style={{ color: payload[0].color }}>
          {payload[0].value}{!isStatus ? 'h' : ''}
        </p>
      </div>
    )
  }
  return null
}

const StatChart = ({ title, icon, data, dataKey, color, year, isMonths, isStatus, isLoading }: StatChartProps) => {
  const chartData = data.slice(0, 12);
  const total = data.reduce((sum, item) => sum + (item[dataKey as keyof DataType] as number || 0), 0);
  
  return (
    <div className={`stat-card ${isLoading ? 'is-loading' : ''}`}>
      <h2>
        {icon} {title} 
        <div className="card-tags">
          <span className="total-tag">Total: {total}{!isStatus ? 'h' : ''}</span>
          <span className="year-tag">{year}</span>
        </div>
      </h2>
      <div className="chart-container">
        {isLoading && (
          <div className="chart-loading-overlay">
            <Loader2 className="animate-spin" size={24} />
            <span>Refreshing...</span>
          </div>
        )}
        <ResponsiveContainer width="100%" height="100%">
          <ReBarChart 
            data={chartData} 
            layout={isMonths ? 'horizontal' : 'vertical'}
            margin={{ left: 10, right: 60, top: 20, bottom: 10 }}
          >
            <XAxis 
              type={isMonths ? 'category' : 'number'} 
              dataKey={isMonths ? 'title' : undefined}
              hide={!isMonths} 
              stroke="#94a3b8" 
              fontSize={14}
              tick={{ fill: '#94a3b8', fontSize: 13, fontWeight: 500 }}
              tickFormatter={t => isMonths ? t.substring(0, 3) : t}
            />
            <YAxis 
              dataKey={isMonths ? undefined : 'title'} 
              type={isMonths ? 'number' : 'category'} 
              width={isMonths ? 50 : 160} 
              stroke="#94a3b8" 
              fontSize={14}
              tick={{ fill: '#94a3b8', fontSize: 13, fontWeight: 500 }}
              tickFormatter={(t) => typeof t === 'string' && t.length > 16 ? t.substring(0, 16) : t}
            />
            <Tooltip content={<CustomTooltip isStatus={isStatus} />} cursor={{ fill: 'rgba(255,255,255,0.05)' }} />
            <Bar 
              dataKey={dataKey} 
              fill={color} 
              radius={isMonths ? [2, 2, 0, 0] : [0, 2, 2, 0]} 
            >
              {!isLoading && (
                <LabelList 
                  dataKey={dataKey} 
                  position={isMonths ? 'top' : 'right'} 
                  offset={15} 
                  fill="#ffffff" 
                  fontSize={14} 
                  fontWeight={700} 
                  formatter={(v: RenderableText) => isStatus ? v : `${v}h`} 
                />
              )}
            </Bar>
          </ReBarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

function App() {
  const [year, setYear] = useState(2024)
  const [compareYear, setCompareYear] = useState<number | null>(null)
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [compareStats, setCompareStats] = useState<StatsResponse | null>(null)
  const [loadingBase, setLoadingBase] = useState(false)
  const [loadingCompare, setLoadingCompare] = useState(false)
  const [viewMode, setViewMode] = useState<'interactive' | 'export'>('interactive')
  const [orientation, setOrientation] = useState<'vertical' | 'horizontal'>('vertical')

  const fromYear = 2021;
  const currentYear = new Date().getFullYear();
  const years = Array.from({ length: currentYear - fromYear + 1 }, (_, i) => fromYear + i)

  useEffect(() => {
    const fetchBase = async () => {
      setLoadingBase(true)
      try {
        const data = await getStats(year)
        setStats(data)
      } catch (err) {
        console.error("Failed to fetch base stats", err)
      } finally {
        setLoadingBase(false)
      }
    }
    fetchBase()
  }, [year])

  useEffect(() => {
    if (!compareYear) {
      setCompareStats(null)
      setLoadingCompare(false)
      return
    }

    const fetchCompare = async () => {
      setLoadingCompare(true)
      try {
        const data = await getStats(compareYear)
        setCompareStats(data)
      } catch (err) {
        console.error("Failed to fetch compare stats", err)
      } finally {
        setLoadingCompare(false)
      }
    }
    fetchCompare()
  }, [compareYear])

  const chartTypes = [
    { id: 'games', title: 'Most Played Games', icon: <Gamepad2 size={24} />, dataKey: 'most_played_games' as keyof StatsResponse, color: '#fca311' },
    { id: 'consoles', title: 'Most Played Consoles', icon: <Monitor size={24} />, dataKey: 'most_played_consoles' as keyof StatsResponse, color: '#fca311' },
    { id: 'platforms', title: 'Most Played Platforms', icon: <Layers size={24} />, dataKey: 'most_played_platforms' as keyof StatsResponse, color: '#fca311' },
    { id: 'series', title: 'Most Played Series', icon: <Sparkles size={24} />, dataKey: 'most_played_series' as keyof StatsResponse, color: '#fca311' },
    { id: 'status', title: 'Games Beaten', icon: <Trophy size={24} />, dataKey: 'games_by_status' as keyof StatsResponse, color: '#B4F8C8' },
    { id: 'months', title: 'Busiest Months', icon: <Calendar size={24} />, dataKey: 'busiest_months' as keyof StatsResponse, color: '#fca311' }
  ]

  return (
    <div className={`app-container ${orientation} ${viewMode}-mode ${compareYear ? 'comparison-mode' : ''}`}>
      <header className="no-screenshot">
        <motion.h1 
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
        >
          GAMER<span>WRAPPED</span>
        </motion.h1>

        <div className="controls">
          <div className="control-group">
            <span className="control-label">Year</span>
            <div className="selector">
              {years.map(y => (
                <button 
                  key={y}
                  className={`btn-small ${year === y ? 'active' : ''}`}
                  onClick={() => {
                    if (y === compareYear) setCompareYear(null);
                    setYear(y);
                  }}
                >
                  {y}
                </button>
              ))}
            </div>
          </div>

          {viewMode === 'interactive' && (
            <div className="control-group">
              <span className="control-label">Compare With</span>
              <div className="selector">
                <button 
                  className={`btn-small ${compareYear === null ? 'active' : ''}`}
                  onClick={() => setCompareYear(null)}
                >
                  None
                </button>
                {years.filter(y => y !== year).map(y => (
                  <button 
                    key={y}
                    className={`btn-small ${compareYear === y ? 'active compare' : ''}`}
                    onClick={() => setCompareYear(y)}
                  >
                    {y}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="control-group">
            <span className="control-label">View</span>
            <div className="selector">
              <button 
                className={`btn-small ${viewMode === 'interactive' ? 'active' : ''}`}
                onClick={() => setViewMode('interactive')}
              >
                <Eye size={14} /> Interactive
              </button>
              <button 
                className={`btn-small ${viewMode === 'export' ? 'active' : ''}`}
                onClick={() => setViewMode('export')}
              >
                <ImageIcon size={14} /> Export
              </button>
            </div>
          </div>

          {viewMode === 'export' && (
            <div className="control-group">
              <span className="control-label">Layout</span>
              <div className="selector">
                <button 
                  className={`btn-small ${orientation === 'vertical' ? 'active' : ''}`}
                  onClick={() => setOrientation('vertical')}
                >
                  Portrait
                </button>
                <button 
                  className={`btn-small ${orientation === 'horizontal' ? 'active' : ''}`}
                  onClick={() => setOrientation('horizontal')}
                >
                  Landscape
                </button>
              </div>
            </div>
          )}
        </div>
      </header>

      <main>
        <AnimatePresence mode="wait">
          {viewMode === 'interactive' ? (
            <motion.div 
              key="interactive"
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="dashboard-grid"
            >
              {chartTypes.map(chart => (
                <div key={chart.id} className="chart-pair">
                  <StatChart 
                    title={chart.title}
                    icon={chart.icon}
                    data={(stats?.[chart.dataKey] as DataType[]) || []}
                    dataKey={chart.id === 'status' ? 'count' : 'playtime'}
                    color={chart.color}
                    year={year}
                    isMonths={chart.id === 'months'}
                    isStatus={chart.id === 'status'}
                    isLoading={loadingBase}
                  />
                  {compareYear && (
                    <StatChart 
                      title={chart.title}
                      icon={chart.icon}
                      data={(compareStats?.[chart.dataKey] as DataType[]) || []}
                      dataKey={chart.id === 'status' ? 'count' : 'playtime'}
                      color="var(--compare-color)"
                      year={compareYear}
                      isMonths={chart.id === 'months'}
                      isStatus={chart.id === 'status'}
                      isLoading={loadingCompare}
                    />
                  )}
                </div>
              ))}
            </motion.div>
          ) : (
            <motion.div 
              key="export"
              initial={{ opacity: 0, scale: 0.98 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 1.02 }}
              className={`charts-layout ${orientation}`}
            >
              {chartTypes.map((chart) => (
                <div key={chart.id} className="export-chart-wrapper">
                  <div className="export-chart-header no-screenshot">
                    <span>{chart.title}</span>
                    <a href={getChartUrl(chart.id, year, orientation)} target="_blank" rel="noreferrer">
                      <Download size={14} />
                    </a>
                  </div>
                  <img 
                    src={getChartUrl(chart.id, year, orientation)} 
                    alt={chart.title}
                    className="wrapped-image"
                  />
                </div>
              ))}
            </motion.div>
          )}
        </AnimatePresence>
      </main>
      <footer className="footer">
        <p>
          {viewMode === 'interactive' ? 
            "Explore and compare my gaming years with interactive charts." : 
            "Snapshot-ready view. Use high-fidelity images for sharing!"
          }
        </p>
        <p>
          Made with <Heart className="heart-icon" /> by{' '}
          <a href="https://aviebrantz.com" target="_blank" rel="noopener noreferrer">
            aviebrantz.com
          </a>
        </p>
      </footer>
    </div>
  )
}

export default App
