import DashboardLayout from '../components/Dashboard/Layout/DashboardLayout.vue'
// GeneralViews
import NotFound from '../components/GeneralViews/NotFoundPage.vue'

// Pages
import Login from 'src/components/Dashboard/Views/Login.vue'
import Practices from 'src/components/Dashboard/Views/Practices.vue'
import Events from 'src/components/Dashboard/Views/Events.vue'
import News from 'src/components/Dashboard/Views/News.vue'

const routes = [
  {
    path: '/',
    component: DashboardLayout,
    redirect: '/news',
    children: [
      {
        path: 'login',
        name: 'Login',
        component: Login
      },
      {
        path: 'news',
        name: 'News',
        component: News
      },
      {
        path: 'practices',
        name: 'Practices',
        component: Practices
      },
      {
        path: 'events',
        name: 'Events',
        component: Events
      }
    ]
  },

  { path: '*', component: NotFound }
]

/**
 * Asynchronously load view (Webpack Lazy loading compatible)
 * The specified component must be inside the Views folder
 * @param  {string} name  the filename (basename) of the view to load.
function view(name) {
   var res= require('../components/Dashboard/Views/' + name + '.vue');
   return res;
};**/

export default routes
