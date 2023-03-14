var {{ prefix }}_options = {
    series: {{ series }},
    chart: {
    height: 350,
    type: 'line',
    zoom: {
      enabled: false
    },
  },
  dataLabels: {
    enabled: false
  },
  stroke: {
    curve: 'smooth',
  },
  {% if chart_type == 'gazoline' %}
  colors: [ '#eb2e2e', '#C70039', '#900C3F' ],
  {% endif %}
  title: {
    text: '{{ title }}',
    align: 'left'
  },
  legend: {
    tooltipHoverFormatter: function(val, opts) {
      return val + ' - <strong>' + opts.w.globals.series[opts.seriesIndex][opts.dataPointIndex] + '</strong>'
    }
  },
  markers: {
    size: 0,
    hover: {
      sizeOffset: 6
    }
  },
  xaxis: {
    {{ xaxis }}
  },
  tooltip: {
    y: [
      {
        title: {
          formatter: function (val) {
            return val + ""
          }
        }
      },
      {
        title: {
          formatter: function (val) {
            return val + ""
          }
        }
      },
      {
        title: {
          formatter: function (val) {
            return val;
          }
        }
      }
    ]
  },
  grid: {
    borderColor: '#f1f1f1',
  }
  };

  var {{ prefix }}_chart = new ApexCharts(document.querySelector("#{{ prefix }}_chart"), {{ prefix }}_options);
  {{ prefix }}_chart.render();