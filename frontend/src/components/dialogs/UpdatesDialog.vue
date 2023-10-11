<script setup lang="ts"></script>

<template>
  <Accordion :activeIndex="0">
    <AccordionTab header="10/11/2023">
      <ul>
        <li>
          <strong>
            <Badge value="Additions" severity="success" /> Added a search bar
            for stations.
          </strong>
          <p>
            As requested, you can now search directly for stations instead of
            being forced to use the map. The search bar also works if you're
            selecting a destination station for travel times.
          </p>
        </li>
        <li>
          <strong>
            <Badge value="Changes" /> Made candlestick graphs hidden by default.
          </strong>
          <p>
            A common criticism was that candlestick graphs were (understandably)
            unreadable for your usual person who probably isn't familiar with
            graphs using for finances and trading. They're still helpful for an
            in-depth view of each period's data, so they're now hidden by
            default and can be toggled using the "Toggle Detailed View" label in
            the legend.
          </p>
        </li>
        <li>
          <strong>
            <Badge value="Changes" /> Made the About page more concise.
          </strong>
          <p>
            The whole "what it does" section is now a more concise "limitations"
            section. Less walls of text for everyone outside of these update
            logs.
          </p>
        </li>
        <li>
          <strong>
            <Badge value="Changes" /> Adjusted several tooltips.
          </strong>
          <p>
            The Headways tooltip is now more concise. The Destination tooltip
            was reworded to be simpler. The Period tooltip also now mentions the
            valid input range.
          </p>
        </li>
        <li>
          <strong>
            <Badge value="Fixes" severity="success" /> Fixed a bug where data
            caching fails if we get too many data points from the MBTA.
          </strong>
          <p>
            It just so happened that in rare cases, the MBTA's APIs returned so
            much data that the backend fed the database too many parameters in a
            single statement. Things should be a bit more resilient to this from
            here on out.
          </p>
        </li>
      </ul>
    </AccordionTab>
    <AccordionTab header="9/29/2023"> Initial release. </AccordionTab>
  </Accordion>
</template>
