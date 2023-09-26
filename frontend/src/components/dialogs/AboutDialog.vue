<script setup lang="ts"></script>

<template>
  <p>
    The MBTA Performance Dashboard is an unofficial web app designed to make
    MBTA performance data more accessible. This is independently developed by
    myself (Kenneth Ho) and is otherwise unaffiliated with the MBTA.
  </p>

  <h3>What this does</h3>
  <ul>
    <li>
      <strong
        >Provides a visual interface for the MBTA's performance data.</strong
      >
      <p>
        You can view data for any MBTA station, which includes headways, dwells,
        and travel times between any two stations on the same route. Keep in
        mind that this data is limited to only the subways, and not the commuter
        rail nor buses.
      </p>
    </li>
    <li>
      <strong>Provides various charts for data.</strong>
      <p>
        This uses candlestick and line graphs to render data. While I can't show
        you all the data points due to how laggy the graph would be, this helps
        condense the data while keeping it performant. You can even hide the
        charts you don't want to see.
      </p>
    </li>
    <li>
      <strong>Caches data from the last 30 days.</strong>
      <p>
        This caches the MBTA's data to avoid angering them and minimize data
        request time. Do note that if you're the first person to request data
        for a stop or it hasn't been cached in a while, it'll take a few seconds
        to cache new data. I also do not save any cached data beyond the last 30
        days due to long-term storage concerns.
      </p>
    </li>
  </ul>

  <h3>Contact</h3>
  <p>
    I'm most active on Discord - my tag's kenesu_h. You can also email me at
    <a href="mailto:officialkennyho@gmail.com">officialkennyho@gmail.com</a>.
  </p>

  &nbsp;

  <h3>Why a dashboard?</h3>
  <p>
    <a
      href="https://www.cbsnews.com/boston/news/mbta-ranking-public-transportation-system-country-boston-riders-commute/"
      >Despite apparently being ranked the second best public transit system in
      the United States</a
    >, the MBTA is pretty infamous for its negligent practices and disrepair
    along its routes, resulting in incidents like derailments, track
    closures/slow zones, or more morbidly,
    <a
      href="https://www.cbsnews.com/boston/news/robinson-lalin-red-line-death-lawsuit-mbta"
      >trains dragging people to death</a
    >.
  </p>
  <p>
    Despite all this, it's difficult to conveniently and visually quantify how
    well the MBTA is perfoming at any point in time. The MBTA itself does
    provide
    <a href="https://www.mbta.com/performance-metrics">performance metrics</a>,
    but I find that these graphs are fairly limited and don't offer a lot of
    freedom to users. Luckily, they also provide access to their data through
    their APIs, though without the necessary technical know-how, most people are
    (understandably) unable to view it.
  </p>
  <p>
    As such, this little dashboard here aims to eliminate that need, and instead
    provide a neat interface for data from the
    <a href="https://www.mbta.com/developers/mbta-performance"
      >MBTA Performance API</a
    >.
  </p>
</template>
