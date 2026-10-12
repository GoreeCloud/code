<script>
import {Bar} from 'vue-chartjs';
import {
  Chart,
  Tooltip,
  BarElement,
  CategoryScale,
  LinearScale,
} from 'chart.js';
import {chartJsColors} from '../utils/color.js';

Chart.defaults.color = chartJsColors.text;
Chart.defaults.borderColor = chartJsColors.border;

Chart.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
);

export default {
  components: {Bar},
  props: {
    locale: {
      type: Object,
      required: true,
    },
  },
  data: () => ({
    colors: {
      barColor: 'green',
    },

    // possible keys:
    // * avatar_link: (...)
    // * commits: (...)
    // * home_link: (...)
    // * login: (...)
    // * name: (...)
    activityTopAuthors: window.config.pageData.repoActivityTopAuthors || [],
    i18nCommitActivity: this,
  }),
  mounted() {
    this.init();
  },
  methods: {
    graphPoints() {
      return {
        datasets: [{
          label: this.locale.commitActivity,
          data: this.activityTopAuthors.map((item) => item.commits),
          backgroundColor: this.colors.barColor,
          barThickness: 40,
          borderWidth: 0,
          tension: 0.3,
        }],
        labels: this.activityTopAuthors.map((item) => item.name),
      };
    },
    getOptions() {
      return {
        responsive: true,
        maintainAspectRatio: false,
        animation: true,
        scales: {
          x: {
            type: 'category',
            grid: {
              display: false,
            },
            ticks: {
              // Disable the drawing of the labels on the x-asis and force them all
              // of them to be 'shown', this avoids them being internally skipped
              // for some data points. We rely on the internally generated ticks
              // to know where to draw our own ticks. Set rotation to 90 degree
              // and disable autoSkip. autoSkip is disabled to ensure no ticks are
              // skipped and rotation is set to avoid messing with the width of the chart.
              color: 'transparent',
              minRotation: 90,
              maxRotation: 90,
              autoSkip: false,
            },
          },
          y: {
            ticks: {
              stepSize: 1,
            },
          },
        },
        plugins: {
          tooltip: {
            intersect: false,
          },
        },
      };
    },
    init() {
      const refStyle = window.getComputedStyle(this.$refs.style);
      this.colors.barColor = refStyle.backgroundColor;

      for (const item of this.activityTopAuthors) {
        const img = new Image();
        img.src = item.avatar_link;
        item.avatar_img = img;
      }

      Chart.register({
        id: 'image_label',
        afterDraw: (chart) => {
          const xAxis = chart.boxes[0];
          const yAxis = chart.boxes[1];
          for (const [index] of xAxis.ticks.entries()) {
            const x = xAxis.getPixelForTick(index);
            const img = this.activityTopAuthors[index].avatar_img;

            chart.ctx.save();
            const [width, height, dx, dy] = this.calcImageSizeAndShift(img);
            chart.ctx.drawImage(img, 0, 0, img.naturalWidth, img.naturalHeight, x - 10 + dx, yAxis.bottom + 10 + dy, width, height);
            chart.ctx.restore();
          }
        },
        beforeEvent: (chart, args) => {
          const event = args.event;
          if (event.type !== 'mousemove' && event.type !== 'click') return;

          const yAxis = chart.boxes[1];
          if (event.y < yAxis.bottom + 10 || event.y > yAxis.bottom + 30) {
            chart.canvas.style.cursor = '';
            return;
          }

          const xAxis = chart.boxes[0];
          const pointIdx = xAxis.ticks.findIndex((_, index) => {
            const x = xAxis.getPixelForTick(index);
            return event.x >= x - 10 && event.x <= x + 10;
          });

          if (pointIdx === -1) {
            chart.canvas.style.cursor = '';
            return;
          }

          chart.canvas.style.cursor = 'pointer';
          if (event.type === 'click' && this.activityTopAuthors[pointIdx].home_link) {
            window.location.href = this.activityTopAuthors[pointIdx].home_link;
          }
        },
      });
    },
    calcImageSizeAndShift(img) {
      const targetSize = 20;
      const [imgWidth, imgHeight] = [img.naturalWidth, img.naturalHeight];

      // The image should be contained in a square,
      // so the scale depends on the longer dimension.
      const scale = targetSize / (Math.max(imgWidth, imgHeight));
      const calcScale = (size) => size * scale;
      const [width, height] = [calcScale(imgWidth), calcScale(imgHeight)];

      // The image should be centered in the 20x20 square.
      const calcShift = (size) => (targetSize - size) / 2;
      const [dx, dy] = [calcShift(width), calcShift(height)];

      return [width, height, dx, dy];
    },
  },
};
</script>
<template>
  <div>
    <div class="activity-bar-graph" ref="style" style="width: 0; height: 0;"/>
    <Bar height="150px" :data="graphPoints()" :options="getOptions()"/>
  </div>
</template>
