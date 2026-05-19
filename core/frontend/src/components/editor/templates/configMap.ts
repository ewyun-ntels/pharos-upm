export const configMap = {
  apiVersion: 'v1',
  kind: 'ConfigMap',
  metadata: {
    name: 'game-demo',
  },
  data: {
    player_initial_lives: '3',
    ui_properties_file_name: 'user-interface.properties',
    'game.properties': 'enemy.types=aliens,monsters\nplayer.maximum-lives=5\n',
    'user-interface.properties': 'color.good=purple\ncolor.bad=yellow\nallow.textmode=true   \n',
  },
};
