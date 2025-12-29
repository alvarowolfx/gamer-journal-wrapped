import axios from 'axios';

export interface MostPlayedByPlaytime {
    title: string;
    playtime: number;
    count: number;
}

export interface MostPlayedGame {
    title: string;
    platform: string;
    console: string;
    playtime: number;
}

export interface MostPlayedByNumGames {
    title: string;
    playtime: number;
    count: number;
}

export interface StatsResponse {
    year: number;
    most_played_consoles: MostPlayedByPlaytime[];
    most_played_platforms: MostPlayedByPlaytime[];
    most_played_games: MostPlayedGame[];
    most_played_series: MostPlayedByPlaytime[];
    games_by_status: MostPlayedByNumGames[];
    busiest_months: MostPlayedByPlaytime[];
}

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const getStats = async (year: number): Promise<StatsResponse> => {
    const response = await axios.get<StatsResponse>(`${API_URL}/api/stats?year=${year}`);
    return response.data;
};

export const getChartUrl = (type: string, year: number, orientation: 'vertical' | 'horizontal' = 'vertical') => {
    return `${API_URL}/api/charts/${type}?year=${year}&orientation=${orientation}`;
};
