<script setup lang="ts"></script>

<template>
  <p>
    The MBTA Performance Dashboard is an unofficial web app designed to make
    MBTA performance data more accessible. This is independently developed by
    Kenneth Ho and is otherwise unaffiliated with the MBTA.
  </p>

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
    And yet, it's difficult to conveniently and visually quantify how well the
    MBTA is perfoming at any given time. The MBTA itself does provide
    <a href="https://www.mbta.com/performance-metrics">performance metrics</a>,
    but I find that these graphs are fairly limited and don't offer a lot of
    freedom to users. Luckily, they also provide access to their data through
    their APIs, though without the necessary technical know-how, most people are
    (understandably) unable to take advantage of it.
  </p>
  <p>
    As such, this little dashboard here aims to eliminate that need and provide
    a neat interface for data from the
    <a href="https://www.mbta.com/developers/mbta-performance"
      >MBTA Performance API</a
    >.
  </p>

  <h3>Limitations</h3>
  <ul>
    <li>Data is limited only to the subways.</li>
    <li>
      To avoid lag, averages are calculated for editable time periods instead of
      showing all points.
    </li>
    <li>
      To avoid the ire of the MBTA, painfully long loading times, and prevent my
      server's local storage from getting unbearably huge, I only want to
      retrieve up to 30 days' worth of data.
    </li>
  </ul>

  <h3>Contact</h3>
  <p>
    I'm most active on Discord - my tag's kenesu_h. You can also email me at
    <a href="mailto:officialkennyho@gmail.com">officialkennyho@gmail.com</a>.
  </p>
</template>
