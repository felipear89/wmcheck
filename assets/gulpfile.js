// REQUIRES, requisições dos plugins
var gulp 		= 		require( 'gulp' )					,
prefix		=		require( 'gulp-autoprefixer' )		,		// https://www.npmjs.com/package/gulp-autoprefixer
sass 		= 		require( 'gulp-sass' )				,		// https://www.npmjs.com/package/gulp-sass
js 			=		require( 'gulp-uglify' )			,		// https://www.npmjs.com/package/gulp-uglify
babel		=		require( 'gulp-babel' )				,		// https://www.npmjs.com/package/gulp-babel
beautify	=		require( 'gulp-jsbeautifier' ) 		,		// https://www.npmjs.com/package/gulp-jsbeautifier
img 		=	 	require( 'gulp-imagemin' )			;		// https://www.npmjs.com/package/gulp-imagemin
//objeto com os caminhos de fonte e destino
var paths = {
  src : {
    sass 	: 	'src/scss/**/*.scss'	,
    js 		: 	'src/js/**/*.js'		,
    img		: 	'src/img/**/*'
  } ,
  dest : {
    sass 	: 	'dist/css'		,
    js 		: 	'dist/js'		,
    img		: 	'dist/img'	,
    beauty 	: 	{
      sass: 'dist/beauty/css' ,
      js 	: 'dist/beauty/js'	
  }
}
}
// taks para compilar o sass, comprimir e colocar os prefixos
gulp.task( 'sass' , function(){
return gulp.src( paths.src.sass )
    .pipe( sass( { outputStyle: 'compressed' } ).on( 'error' , sass.logError ) )
    .pipe( prefix( { browsers: ['last 10 versions'] } ) )
    .pipe( gulp.dest( paths.dest.sass ) );
});
// taks de minificar e executar babel (retirar de ES6 para ES5) o js
gulp.task( 'js' , function(){
return gulp.src( paths.src.js )
    .pipe( babel( { presets : [ 'es2015' ] } ) )
    .pipe( js() )
    .pipe( gulp.dest( paths.dest.js ) );
});
// task de comprimir imagens
gulp.task( 'img' , function(){
return gulp.src( paths.src.img )
    .pipe( img( { optimizationLevel: 5 , progressive: true  } ) )
    .pipe( gulp.dest( paths.dest.img ) );
});
// task watch para assistir dependencias
gulp.task( 'watch' , function(){
gulp.watch( paths.src.sass , ['sass'] );
gulp.watch( paths.src.js , ['js'] );
gulp.watch( paths.src.img , ['img'] );
});
// task padrao que executa watch (assiste todos os arquivos)
gulp.task( 'default' , ['watch'] , function(){
// gulp.start( 'sass' , 'js' , 'img' );
gulp.start( 'sass' , 'img' );
});
// task para beautify (descomprimir) css
gulp.task( 'beauty-css' , function(){
return gulp.src( paths.dest.sass + '/**/*.css' )
    .pipe( beautify() )
    .pipe( gulp.dest( paths.dest.beauty.js ) );
});
// task para beautify (descomprimir) js
gulp.task( 'beauty-js' , function(){
return gulp.src( paths.dest.js + '/**/*.js' )
    .pipe( beautify() )
    .pipe( gulp.dest( paths.dest.beauty.js ) );
});