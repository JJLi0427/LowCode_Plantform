<template>
  <div id="app">
    <h1>Computer Traffic Monitoring Tool</h1>
    <div class="monitor">
      <div class="section">
        <h2>Real-Time Rates</h2>
        <p>Upload: {{ uploadRate }} Mbps</p>
        <p>Download: {{ downloadRate }} Mbps</p>
      </div>
      <div class="section">
        <h2>Usage Statistics</h2>
        <p>Total Uploaded: {{ totalUpload }} MB</p>
        <p>Total Downloaded: {{ totalDownload }} MB</p>
      </div>
    </div>
    <div class="charts">
      <div class="chart-container">
        <h2>Upload Rate</h2>
        <canvas id="uploadCanvas" width="400" height="200"></canvas>
      </div>
      <div class="chart-container">
        <h2>Download Rate</h2>
        <canvas id="downloadCanvas" width="400" height="200"></canvas>
      </div>
      <div class="chart-container">
        <h2>Distribution <span style="color: red;">download</span>:<span style="color: blue;">upload</span></h2>
        <canvas id="pieCanvas" width="200" height="200"></canvas>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      uploadRate: 0, downloadRate: 0, totalUpload: 0, totalDownload: 0,
      timer: null, maxDataPoints: 20, uploadHistory: [], downloadHistory: [],
      uploadCtx: null, downloadCtx: null, canvasWidth: 400, canvasHeight: 200,
      pieCtx: null, pieCanvasWidth: 200, pieCanvasHeight: 200
    };
  },
  methods: {
    updateTraffic() {
      const up = parseFloat((Math.random() * 10).toFixed(2));
      const down = parseFloat((Math.random() * 10).toFixed(2));
      this.uploadRate = up;
      this.downloadRate = down;
      this.totalUpload = parseFloat((parseFloat(this.totalUpload) + up).toFixed(2));
      this.totalDownload = parseFloat((parseFloat(this.totalDownload) + down).toFixed(2));
      
      this.uploadHistory.push(up);
      this.downloadHistory.push(down);
      if (this.uploadHistory.length > this.maxDataPoints) {
        this.uploadHistory.shift();
      }
      if (this.downloadHistory.length > this.maxDataPoints) {
        this.downloadHistory.shift();
      }
      
      this.drawChart(this.uploadCtx, this.uploadHistory, 'blue');
      this.drawChart(this.downloadCtx, this.downloadHistory, 'red');

      this.drawPieChart();
    },
    drawChart(ctx, data, strokeStyle) {
      ctx.clearRect(0, 0, this.canvasWidth, this.canvasHeight);
      this.drawAxes(ctx);
      
      const gap = this.canvasWidth / (this.maxDataPoints - 1);
      const maxValue = Math.max(...data, 10);
      
      ctx.beginPath();
      ctx.strokeStyle = strokeStyle;
      data.forEach((value, index) => {
        const x = index * gap;
        const y = this.canvasHeight - (value / maxValue) * this.canvasHeight;
        if (index === 0) {
          ctx.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }
      });
      ctx.stroke();
    },
    drawAxes(ctx) {
      ctx.beginPath();
      ctx.strokeStyle = '#000';
      ctx.moveTo(0, this.canvasHeight);
      ctx.lineTo(this.canvasWidth, this.canvasHeight);
      ctx.stroke();

      ctx.beginPath();
      ctx.moveTo(0, 0);
      ctx.lineTo(0, this.canvasHeight);
      ctx.stroke();

      ctx.font = "12px sans-serif";
      ctx.fillStyle = '#000';
      const xLabel = "Time (s)";
      const xLabelWidth = ctx.measureText(xLabel).width;
      ctx.fillText(xLabel, (this.canvasWidth - xLabelWidth) / 2, this.canvasHeight - 5);

      const yLabel = "Rate (Mbps)";
      ctx.save();
      ctx.translate(15, this.canvasHeight / 2 + ctx.measureText(yLabel).width / 2);
      ctx.rotate(-Math.PI / 2);
      ctx.fillText(yLabel, 0, 0);
      ctx.restore();
    },
    drawPieChart() {
      this.pieCtx.clearRect(0, 0, this.pieCanvasWidth, this.pieCanvasHeight);

      const upload = this.totalUpload;
      const download = this.totalDownload;
      const total = upload + download;
      const centerX = this.pieCanvasWidth / 2;
      const centerY = this.pieCanvasHeight / 2;
      const radius = Math.min(this.pieCanvasWidth, this.pieCanvasHeight) / 2 - 10; // 留出边距

      if (total === 0) {
        this.pieCtx.fillStyle = '#ccc';
        this.pieCtx.fillRect(0, 0, this.pieCanvasWidth, this.pieCanvasHeight);
        return;
      }

      let startAngle = 0;
      const uploadAngle = (upload / total) * 2 * Math.PI;
      this.pieCtx.beginPath();
      this.pieCtx.fillStyle = 'blue';
      this.pieCtx.moveTo(centerX, centerY);
      this.pieCtx.arc(centerX, centerY, radius, startAngle, startAngle + uploadAngle);
      this.pieCtx.closePath();
      this.pieCtx.fill();

      startAngle += uploadAngle;
      const downloadAngle = (download / total) * 2 * Math.PI;
      this.pieCtx.beginPath();
      this.pieCtx.fillStyle = 'red';
      this.pieCtx.moveTo(centerX, centerY);
      this.pieCtx.arc(centerX, centerY, radius, startAngle, startAngle + downloadAngle);
      this.pieCtx.closePath();
      this.pieCtx.fill();

      this.pieCtx.fillStyle = '#fff';
      this.pieCtx.font = 'bold 14px sans-serif';
      const uploadPct = ((upload / total) * 100).toFixed(1) + '%';
      const downloadPct = ((download / total) * 100).toFixed(1) + '%';
      this.pieCtx.fillText(uploadPct, centerX - radius / 2, centerY);
      this.pieCtx.fillText(downloadPct, centerX + radius / 4, centerY);
    }
  },
  mounted() {
    this.uploadCtx = document.getElementById('uploadCanvas').getContext('2d');
    this.downloadCtx = document.getElementById('downloadCanvas').getContext('2d');
    this.pieCtx = document.getElementById('pieCanvas').getContext('2d');
    
    this.timer = setInterval(this.updateTraffic, 1000);
  },
  beforeDestroy() {
    clearInterval(this.timer);
  }
};
</script>

<style scoped>
#app {
  text-align: center;
  margin: 20px;
}
.monitor {
  padding: 20px;
  background: #f0f0f0;
  border-radius: 5px;
}
.section {
  margin: 20px 0;
}
p {
  margin: 5px 0;
  font-size: 1.2em;
}
.charts {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-around;
  margin-top: 20px;
}
.chart-container {
  margin-bottom: 20px;
}
</style>