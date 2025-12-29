package imagegen

const (
	QueryMostPlayedConsoles = `
		select c.name as title, ROUND(sum(p.playtime)/(60*60), 0) as playtime, count(*) as count
		from playthroughs p
			inner join consoles c on JSON_CONTAINS(p.console, CONCAT('"', c.record_id, '"'))
		where p.year_start_date = ?
		group by c.name
		order by playtime desc`

	QueryMostPlayedPlatforms = `
		select pt.name as title, ROUND(sum(p.playtime)/(60*60), 0) as playtime, count(*) as count
		from playthroughs p	
			inner join games g on JSON_CONTAINS(p.games, CONCAT('"', g.record_id, '"'))
			inner join platforms pt on JSON_CONTAINS(g.platforms, CONCAT('"', pt.record_id, '"'))
		where p.year_start_date = ? 
		group by pt.name
		order by playtime desc;`

	QueryMostPlayedGames = `
		select g.name as title, pt.name as platform, c.name as console, ROUND(p.playtime/(60*60), 0) as playtime
		from playthroughs p	
			inner join games g on JSON_CONTAINS(p.games, CONCAT('"', g.record_id, '"'))
			inner join consoles c on JSON_CONTAINS(p.console, CONCAT('"', c.record_id, '"'))
			inner join platforms pt on JSON_CONTAINS(g.platforms, CONCAT('"', pt.record_id, '"'))
		where p.year_start_date = ?
			and p.status not in ('Abandoned')
		order by playtime desc;`

	QueryMostPlayedSeries = `
		select s.name as title, ROUND(sum(p.playtime)/(60*60), 0) as playtime, count(*) as count
		from playthroughs p	
			inner join games g on JSON_CONTAINS(p.games, CONCAT('"', g.record_id, '"'))
			inner join serie s on JSON_CONTAINS(g.serie, CONCAT('"', s.record_id, '"'))
		where p.year_start_date = ?
		group by s.name
		order by playtime desc;`

	QueryGamesByStatus = `
		select p.status as title, ROUND(sum(p.playtime)/(60*60), 0) as playtime, count(*) as count
		from playthroughs p	
		where p.year_start_date = ?
			and p.status not in ('Playing')
		group by p.status
		order by count desc;`

	QueryBusiestMonths = `
		select EXTRACT(MONTH from p.start_date) as title, ROUND(sum(p.playtime)/(60*60), 0) as playtime, count(*) as count
			from playthroughs p	
		where p.year_start_date = ?
		group by title
		order by title asc;`
)
