import DashboardLayout from '../components/Dashboard/Layout/DashboardLayout.vue'
// GeneralViews
import NotFound from '../components/GeneralViews/NotFoundPage.vue'

// Pages
import Login from 'src/components/Dashboard/Views/Login.vue'
import Initialize from 'src/components/Dashboard/Views/Initialize.vue'
import Members from 'src/components/Dashboard/Views/Members.vue'
import MemberEdit from 'src/components/Dashboard/Views/MemberEdit.vue'
import Practices from 'src/components/Dashboard/Views/Practices.vue'
import PracticeEdit from 'src/components/Dashboard/Views/PracticeEdit.vue'
import Events from 'src/components/Dashboard/Views/Events.vue'

const routes = [
  {
    path: '/',
    component: DashboardLayout,
    redirect: '/initialize',
    children: [
      {
        path: 'login',
        name: 'Login',
        component: Login
      },
      {
        path: 'initialize',
        name: 'Initialize',
        component: Initialize
      },
      {
        path: 'members',
        name: 'Members',
        component: Members
      },
      {
        path: 'memberEdit',
        name: 'MemberAdd',
        component: MemberEdit
      },
      {
        path: 'memberEdit/:uuid',
        name: 'MemberEdit',
        component: MemberEdit
      },
      {
        path: 'practices',
        name: 'Practices',
        component: Practices
      },
      {
        path: 'practiceEdit',
        name: 'PracticeAdd',
        component: PracticeEdit
      },
      {
        path: 'practiceEdit/:uuid',
        name: 'practiceEdit',
        component: PracticeEdit
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
