<template>
  <div>
    <NavigationBar />
    <b-container id="app">
      <router-view></router-view>
    </b-container>
    <PageFooter />
  </div>
</template>

<script>
import {getRecent} from '@/controllers/Recent.js';
import initializeUserState from '@/controllers/UserState.js';
import {loadUserKit} from '@/controllers/UserKit.js';

import PageFooter from '@/components/PageFooter';
import NavigationBar from '@/components/NavigationBar';

export default {
  name: 'app',
  components: {
    PageFooter,
    NavigationBar,
  },
  created() {
    loadUserKit(process.env.VUE_APP_USERKIT_APP_ID).then((userKit) => {
      if (userKit.isLoggedIn() === true) {
        initializeUserState();
      } else {
        this.$store.commit('clearUserState');
        if (this.routeRequiresLogin) {
          this.$router.push('/login');
        }
      }
    });
    getRecent(/*start=*/ 0).then((recentEntries) => {
      this.$store.commit('setRecent', recentEntries);
    });
  },
  computed: {
    routeRequiresLogin: function () {
      const routeName = this.$router.currentRoute.name;
      if (!routeName) {
        return false;
      }
      if (routeName === 'Preferences') {
        return true;
      }
      if (routeName.indexOf('Edit') === 0) {
        return true;
      }
      return false;
    },
  },
};
</script>

<style>
@import '~@fortawesome/fontawesome-svg-core/styles.css';

#app {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  margin-top: 60px;
}

#app a.btn {
  color: white;
}

/* TODO: Move these to the view entry component */
#app a.page-link {
  color: white;
}

#app .page-link {
  border: 1px solid rgb(124, 133, 145);
}
</style>
